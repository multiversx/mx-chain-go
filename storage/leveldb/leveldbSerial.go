package leveldb

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/closing"
	"github.com/ElrondNetwork/elrond-go/errors"
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
	sizeBatch         int
	batch             storage.Batcher
	mutBatch          sync.RWMutex
	dbAccess          chan serialQueryer
	cancel            context.CancelFunc
	closer            core.SafeCloser
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
		sizeBatch:         0,
		dbAccess:          make(chan serialQueryer),
		cancel:            cancel,
		closer:            closing.NewSafeChanCloser(),
	}

	dbStore.batch = NewBatch()

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
			err := s.putBatch()
			if err != nil {
				log.Warn("leveldb serial putBatch", "error", err.Error())
				continue
			}
		case <-ctx.Done():
			log.Debug("batchTimeoutHandle - closing", "path", s.path)
			return
		}
	}
}

func (s *SerialDB) updateBatchWithIncrement() error {
	s.mutBatch.Lock()
	s.sizeBatch++
	if s.sizeBatch < s.maxBatchSize {
		s.mutBatch.Unlock()
		return nil
	}
	s.mutBatch.Unlock()

	err := s.putBatch()

	return err
}

// Put adds the value to the (key, val) storage medium
func (s *SerialDB) Put(key, val []byte) error {
	if s.isClosed() {
		return errors.ErrDBIsClosed
	}

	s.mutBatch.RLock()
	err := s.batch.Put(key, val)
	s.mutBatch.RUnlock()
	if err != nil {
		return err
	}

	return s.updateBatchWithIncrement()
}

// Get returns the value associated to the key
func (s *SerialDB) Get(key []byte) ([]byte, error) {
	if s.isClosed() {
		return nil, errors.ErrDBIsClosed
	}

	s.mutBatch.RLock()
	if s.batch.IsRemoved(key) {
		s.mutBatch.RUnlock()
		return nil, storage.ErrKeyNotFound
	}

	data := s.batch.Get(key)
	s.mutBatch.RUnlock()

	if data != nil {
		return data, nil
	}

	ch := make(chan *pairResult)
	req := &getAct{
		key:     key,
		resChan: ch,
	}

	err := s.tryWriteInDbAccessChan(req)
	if err != nil {
		return nil, err
	}
	result := <-ch
	close(ch)

	if result.err == leveldb.ErrNotFound {
		return nil, storage.ErrKeyNotFound
	}
	if result.err != nil {
		return nil, result.err
	}

	return result.value, nil
}

// Has returns nil if the given key is present in the persistence medium
func (s *SerialDB) Has(key []byte) error {
	if s.isClosed() {
		return errors.ErrDBIsClosed
	}

	s.mutBatch.RLock()
	if s.batch.IsRemoved(key) {
		s.mutBatch.RUnlock()
		return storage.ErrKeyNotFound
	}

	data := s.batch.Get(key)
	s.mutBatch.RUnlock()

	if data != nil {
		return nil
	}

	ch := make(chan error)
	req := &hasAct{
		key:     key,
		resChan: ch,
	}

	err := s.tryWriteInDbAccessChan(req)
	if err != nil {
		return err
	}
	result := <-ch
	close(ch)

	return result
}

func (s *SerialDB) tryWriteInDbAccessChan(req serialQueryer) error {
	select {
	case s.dbAccess <- req:
		return nil
	case <-s.closer.ChanClose():
		return errors.ErrDBIsClosed
	}
}

// putBatch writes the Batch data into the database
func (s *SerialDB) putBatch() error {
	s.mutBatch.Lock()
	dbBatch, ok := s.batch.(*batch)
	if !ok {
		s.mutBatch.Unlock()
		return storage.ErrInvalidBatch
	}
	s.sizeBatch = 0
	s.batch = NewBatch()
	s.mutBatch.Unlock()

	ch := make(chan error)
	req := &putBatchAct{
		batch:   dbBatch,
		resChan: ch,
	}

	err := s.tryWriteInDbAccessChan(req)
	if err != nil {
		return err
	}
	result := <-ch
	close(ch)

	return result
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
func (s *SerialDB) Remove(key []byte) error {
	if s.isClosed() {
		return errors.ErrDBIsClosed
	}

	s.mutBatch.Lock()
	_ = s.batch.Delete(key)
	s.mutBatch.Unlock()

	return s.updateBatchWithIncrement()
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
	_ = s.putBatch()
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
		case queryer := <-s.dbAccess:
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
