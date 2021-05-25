package leveldb

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var _ storage.Persister = (*SerialDB)(nil)

// SerialDB holds a pointer to the leveldb database and the path to where it is stored.
type SerialDB struct {
	*baseLevelDb
	path              string
	maxBatchSize      int
	batchDelaySeconds int
	sizeBatch         int
	batch             storage.Batcher
	mutBatch          sync.RWMutex
	dbAccess          chan serialQueryer
	cancel            context.CancelFunc
	mutClosed         sync.Mutex
	closed            bool
	closingChan       chan struct{}
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
		db: db,
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbStore := &SerialDB{
		baseLevelDb:       bldb,
		path:              path,
		maxBatchSize:      maxBatchSize,
		batchDelaySeconds: batchDelaySeconds,
		sizeBatch:         0,
		dbAccess:          make(chan serialQueryer),
		cancel:            cancel,
		closed:            false,
		closingChan:       make(chan struct{}),
	}

	dbStore.batch = NewBatch()

	go dbStore.batchTimeoutHandle(ctx)
	go dbStore.processLoop(ctx)

	runtime.SetFinalizer(dbStore, func(db *SerialDB) {
		_ = db.Close()
	})

	log.Debug("opened serial level db persister", "path", path)

	return dbStore, nil
}

func (s *SerialDB) batchTimeoutHandle(ctx context.Context) {
	for {
		select {
		case <-time.After(time.Duration(s.batchDelaySeconds) * time.Second):
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
		return storage.ErrSerialDBIsClosed
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
		return nil, storage.ErrSerialDBIsClosed
	}

	s.mutBatch.RLock()
	data := s.batch.Get(key)
	s.mutBatch.RUnlock()

	if data != nil {
		if bytes.Equal(data, []byte(removed)) {
			return nil, storage.ErrKeyNotFound
		}
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
		return storage.ErrSerialDBIsClosed
	}

	s.mutBatch.RLock()
	data := s.batch.Get(key)
	s.mutBatch.RUnlock()

	if data != nil {
		if bytes.Equal(data, []byte(removed)) {
			return storage.ErrKeyNotFound
		}
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
	case <-s.closingChan:
		return storage.ErrSerialDBIsClosed
	}
}

// Init initializes the storage medium and prepares it for usage
func (s *SerialDB) Init() error {
	// no special initialization needed
	return nil
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
	s.mutClosed.Lock()
	isClosed := s.closed
	s.mutClosed.Unlock()

	return isClosed
}

// Close closes the files/resources associated to the storage medium
func (s *SerialDB) Close() error {
	s.mutClosed.Lock()
	defer s.mutClosed.Unlock()

	if s.closed {
		return nil
	}

	close(s.closingChan)
	s.closed = true
	_ = s.putBatch()

	s.cancel()

	return s.db.Close()
}

// Remove removes the data associated to the given key
func (s *SerialDB) Remove(key []byte) error {
	if s.isClosed() {
		return storage.ErrSerialDBIsClosed
	}

	s.mutBatch.Lock()
	_ = s.batch.Delete(key)
	s.mutBatch.Unlock()

	return s.updateBatchWithIncrement()
}

// Destroy removes the storage medium stored data
func (s *SerialDB) Destroy() error {
	s.mutBatch.Lock()
	s.batch.Reset()
	s.sizeBatch = 0
	s.mutBatch.Unlock()

	s.cancel()

	s.mutClosed.Lock()
	s.closed = true
	err := s.db.Close()
	s.mutClosed.Unlock()

	if err != nil {
		return err
	}

	err = os.RemoveAll(s.path)

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
