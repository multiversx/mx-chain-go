package leveldb

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/closing"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var _ storage.Persister = (*SerialDB)(nil)

// SerialDB holds a pointer to the leveldb database and the path to where it is stored.
type SerialDB struct {
	*baseLevelDb
	maxBatchSize      int
	batchDelaySeconds int
	waitingOperations *waitingOperationsPrinter

	lowPrioBatch  *batchWrapper
	highPrioBatch *batchWrapper

	lowPrioDbAccess  chan serialQueryer
	highPrioDbAccess chan serialQueryer
	cancel           context.CancelFunc
	closer           core.SafeCloser
}

// NewSerialDB is a constructor for the leveldb persister
// It creates the files in the location given as parameter
func NewSerialDB(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (s *SerialDB, err error) {
	err = os.MkdirAll(path, rwxOwner)
	if err != nil {
		return nil, err
	}

	if maxOpenFiles < 1 {
		return nil, storage.ErrInvalidNumOpenFiles
	}

	options := &opt.Options{
		// disable internal cache
		BlockCacheCapacity:     -1,
		OpenFilesCacheCapacity: maxOpenFiles,
	}

	db, err := openLevelDB(path, options)
	if err != nil {
		return nil, fmt.Errorf("%w for path %s", err, path)
	}

	bldb := &baseLevelDb{
		db:   db,
		path: path,
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbStore := &SerialDB{
		baseLevelDb:       bldb,
		maxBatchSize:      maxBatchSize,
		batchDelaySeconds: batchDelaySeconds,
		lowPrioBatch:      newBatchWrapper(),
		highPrioBatch:     newBatchWrapper(),
		lowPrioDbAccess:   make(chan serialQueryer),
		highPrioDbAccess:  make(chan serialQueryer),
		cancel:            cancel,
		closer:            closing.NewSafeChanCloser(),
		waitingOperations: newWaitingOperationsPrinter(),
	}

	go dbStore.batchTimeoutHandle(ctx)
	go dbStore.processLoop(ctx)

	runtime.SetFinalizer(dbStore, func(db *SerialDB) {
		_ = db.Close()
	})

	crtCounter := atomic.AddUint32(&loggingDBCounter, 1)
	log.Debug("opened serial level db persister", "path", path, "created pointer", fmt.Sprintf("%p", bldb.db), "global db counter", crtCounter)

	return dbStore, nil
}

func (s *SerialDB) batchTimeoutHandle(ctx context.Context) {
	interval := time.Duration(s.batchDelaySeconds) * time.Second
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		timer.Reset(interval)

		select {
		case <-timer.C:
			s.putAllBatches()
		case <-ctx.Done():
			log.Debug("batchTimeoutHandle - closing", "path", s.path)
			return
		}
	}
}

func (s *SerialDB) putAllBatches() {
	s.putBatchHandlingErrors(common.HighPriority)
	s.putBatchHandlingErrors(common.LowPriority)
}

func (s *SerialDB) putBatchHandlingErrors(priority common.StorageAccessType) {
	wrappedBatch, writeChan, err := s.getBatchWrapperAndChan(priority)
	if err != nil {
		log.Warn("SerialDB.putBatchHandlingErrors", "priority", priority, "error", err)
		return
	}

	err = s.putBatch(wrappedBatch, writeChan, priority)
	if err != nil {
		log.Warn("SerialDB.putBatchHandlingErrors -> putBatch", "error", err.Error())
		return
	}
}

func (s *SerialDB) updateBatchWithIncrement(wrappedBatch *batchWrapper, writeChan chan serialQueryer, priority common.StorageAccessType) error {
	if wrappedBatch.updateBatchSizeReturningCurrent() < s.maxBatchSize {
		return nil
	}

	return s.putBatch(wrappedBatch, writeChan, priority)
}

func (s *SerialDB) getBatchWrapperAndChan(priority common.StorageAccessType) (*batchWrapper, chan serialQueryer, error) {
	if priority == common.HighPriority {
		return s.highPrioBatch, s.highPrioDbAccess, nil
	}
	if priority == common.LowPriority {
		return s.lowPrioBatch, s.lowPrioDbAccess, nil
	}

	return nil, nil, storage.ErrInvalidPriorityType
}

// Put adds the value to the (key, val) storage medium
func (s *SerialDB) Put(key, val []byte, priority common.StorageAccessType) error {
	if s.isClosed() {
		return storage.ErrDBIsClosed
	}

	wrappedBatch, writeChan, err := s.getBatchWrapperAndChan(priority)
	if err != nil {
		return fmt.Errorf("%w in SerialDB.Put", err)
	}

	err = wrappedBatch.put(key, val)
	if err != nil {
		return err
	}

	return s.updateBatchWithIncrement(wrappedBatch, writeChan, priority)
}

// Get returns the value associated to the key
func (s *SerialDB) Get(key []byte, priority common.StorageAccessType) ([]byte, error) {
	if s.isClosed() {
		return nil, storage.ErrDBIsClosed
	}

	wrappedBatch, writeChan, err := s.getBatchWrapperAndChan(priority)
	if err != nil {
		return nil, fmt.Errorf("%w in SerialDB.Get", err)
	}

	data := wrappedBatch.get(key)
	if data != nil {
		if bytes.Equal(data, []byte(removed)) {
			return nil, storage.ErrKeyNotFound
		}
		return data, nil
	}

	resChan := make(chan *result)
	req := &getAction{
		key:     key,
		resChan: resChan,
	}

	res := s.executeSerialDBRequest(req, writeChan, priority, "get")

	if res.err == leveldb.ErrNotFound {
		return nil, storage.ErrKeyNotFound
	}
	if res.err != nil {
		return nil, res.err
	}

	return res.value, nil
}

// Has returns nil if the given key is present in the persistence medium
func (s *SerialDB) Has(key []byte, priority common.StorageAccessType) error {
	if s.isClosed() {
		return storage.ErrDBIsClosed
	}

	wrappedBatch, writeChan, err := s.getBatchWrapperAndChan(priority)
	if err != nil {
		return fmt.Errorf("%w in SerialDB.Has", err)
	}

	data := wrappedBatch.get(key)
	if data != nil {
		if bytes.Equal(data, []byte(removed)) {
			return storage.ErrKeyNotFound
		}
		return nil
	}

	resChan := make(chan *result)
	req := &hasAction{
		key:     key,
		resChan: resChan,
	}

	return s.executeSerialDBRequest(req, writeChan, priority, "has").err
}

func (s *SerialDB) tryWriteInDbAccessChan(req serialQueryer, ch chan serialQueryer) error {
	select {
	case ch <- req:
		return nil
	case <-s.closer.ChanClose():
		return storage.ErrDBIsClosed
	}
}

func (s *SerialDB) executeSerialDBRequest(req serialQueryer, writeChan chan serialQueryer, priority common.StorageAccessType, operation string) *result {
	s.waitingOperations.startExecuting(priority)
	start := time.Now()

	defer s.waitingOperations.endExecuting(priority, operation, start)

	err := s.tryWriteInDbAccessChan(req, writeChan)
	if err != nil {
		return &result{
			err: err,
		}
	}

	res := <-req.resultChan()
	close(req.resultChan())

	return res
}

// putBatch writes the Batch data into the database
func (s *SerialDB) putBatch(wrappedBatch *batchWrapper, writeChan chan serialQueryer, priority common.StorageAccessType) error {
	lastBatch := wrappedBatch.resetBatchReturningLast()
	dbBatch, ok := lastBatch.(*batch)
	if !ok {
		return storage.ErrInvalidBatch
	}

	resChan := make(chan *result)
	req := &putBatchAction{
		batch:   dbBatch,
		resChan: resChan,
	}

	return s.executeSerialDBRequest(req, writeChan, priority, "put").err
}

func (s *SerialDB) isClosed() bool {
	db := s.getDbPointer()

	return db == nil
}

// Close closes the files/resources associated to the storage medium
func (s *SerialDB) Close() error {
	// calling close on the SafeCloser instance should be the last instruction called
	// (just to close some go routines started as edge cases that would otherwise hang)
	defer s.closer.Close()

	return s.doClose()
}

// Remove removes the data associated to the given key
func (s *SerialDB) Remove(key []byte, priority common.StorageAccessType) error {
	if s.isClosed() {
		return storage.ErrDBIsClosed
	}

	wrappedBatch, writeChan, err := s.getBatchWrapperAndChan(priority)
	if err != nil {
		return fmt.Errorf("%w in SerialDB.Remove", err)
	}

	wrappedBatch.delete(key)

	return s.updateBatchWithIncrement(wrappedBatch, writeChan, priority)
}

// Destroy removes the storage medium stored data
func (s *SerialDB) Destroy() error {
	log.Debug("serialDB.Destroy", "path", s.path)

	// calling close on the SafeCloser instance should be the last instruction called
	// (just to close some go routines started as edge cases that would otherwise hang)
	defer s.closer.Close()

	err := s.doClose()
	if err == nil {
		return os.RemoveAll(s.path)
	}

	return err
}

// DestroyClosed removes the already closed storage medium stored data
func (s *SerialDB) DestroyClosed() error {
	err := os.RemoveAll(s.path)
	if err != nil {
		log.Error("error destroy closed", "error", err, "path", s.path)
	}
	return err
}

// doClose will handle the closing of the internal components
// must be called under mutex protection
// TODO: re-use this function in leveldb.go as well
func (s *SerialDB) doClose() error {
	s.putAllBatches()
	s.cancel()

	db := s.makeDbPointerNilReturningLast()
	if db != nil {
		return db.Close()
	}

	return nil
}

func (s *SerialDB) processLoop(ctx context.Context) {
	for {
		select {
		case queryer := <-s.highPrioDbAccess:
			queryer.request(s)
			continue
		case <-ctx.Done():
			log.Debug("processLoop - closing the leveldb process loop", "path", s.path)
			return
		default:
		}

		select {
		case queryer := <-s.highPrioDbAccess:
			queryer.request(s)
		case queryer := <-s.lowPrioDbAccess:
			queryer.request(s)
		case <-ctx.Done():
			log.Debug("processLoop - closing the leveldb process loop", "path", s.path)
			return
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SerialDB) IsInterfaceNil() bool {
	return s == nil
}
