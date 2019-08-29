package leveldb

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// SerialDB holds a pointer to the leveldb database and the path to where it is stored.
type SerialDB struct {
	db                *leveldb.DB
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

	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbStore := &SerialDB{
		db:                db,
		path:              path,
		maxBatchSize:      maxBatchSize,
		batchDelaySeconds: batchDelaySeconds,
		sizeBatch:         0,
		dbAccess:          make(chan serialQueryer),
		cancel:            cancel,
		closed:            false,
	}

	dbStore.batch = NewBatch()

	go dbStore.batchTimeoutHandle(ctx)
	go dbStore.processLoop(ctx)

	return dbStore, nil
}

func (s *SerialDB) batchTimeoutHandle(ctx context.Context) {
	ct, _ := context.WithCancel(ctx)

	for {
		select {
		case <-time.After(time.Duration(s.batchDelaySeconds) * time.Second):
			err := s.putBatch()
			if err != nil {
				log.Error(err.Error())
				continue
			}
		case <-ct.Done():
			log.Info("closing the timed batch handler")
			return
		}
	}
}

// Put adds the value to the (key, val) storage medium
func (s *SerialDB) Put(key, val []byte) error {
	if s.isClosed() {
		return storage.ErrSerialDBIsClosed
	}

	s.mutBatch.Lock()
	err := s.batch.Put(key, val)
	if err != nil {
		s.mutBatch.Unlock()
		return err
	}

	s.sizeBatch++
	if s.sizeBatch < s.maxBatchSize {
		s.mutBatch.Unlock()
		return nil
	}
	s.mutBatch.Unlock()

	err = s.putBatch()

	return err
}

// Get returns the value associated to the key
func (s *SerialDB) Get(key []byte) ([]byte, error) {
	if s.isClosed() {
		return nil, storage.ErrSerialDBIsClosed
	}

	ch := make(chan *pairResult)
	req := &getAct{
		key:     key,
		resChan: ch,
	}

	s.dbAccess <- req
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

// Has returns true if the given key is present in the persistence medium
func (s *SerialDB) Has(key []byte) error {
	if s.isClosed() {
		return storage.ErrSerialDBIsClosed
	}

	ch := make(chan error)
	req := &hasAct{
		key:     key,
		resChan: ch,
	}

	s.dbAccess <- req
	result := <-ch
	close(ch)

	return result
}

// Init initializes the storage medium and prepares it for usage
func (s *SerialDB) Init() error {
	// no special initialization needed
	return nil
}

// putBatch writes the Batch data into the database
func (s *SerialDB) putBatch() error {
	s.mutBatch.Lock()
	batch, ok := s.batch.(*batch)
	if !ok {
		s.mutBatch.Unlock()
		return storage.ErrInvalidBatch
	}
	s.sizeBatch = 0
	s.batch = NewBatch()
	s.mutBatch.Unlock()

	ch := make(chan error)
	req := &putBatchAct{
		batch:   batch,
		resChan: ch,
	}

	s.dbAccess <- req
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
	if s.closed {
		s.mutClosed.Unlock()
		return nil
	}
	s.closed = true
	s.mutClosed.Unlock()

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

	ch := make(chan error)
	req := &delAct{
		key:     key,
		resChan: ch,
	}

	s.dbAccess <- req
	result := <-ch
	close(ch)

	return result
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
	s.mutClosed.Unlock()

	err := s.db.Close()
	if err != nil {
		return err
	}

	err = os.RemoveAll(s.path)

	return err
}

func (s *SerialDB) processLoop(ctx context.Context) {
	ct, _ := context.WithCancel(ctx)

	for {
		select {
		case queryer := <-s.dbAccess:
			queryer.request(s)
		case <-ct.Done():
			log.Info("closing the leveldb process loop")
			return
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SerialDB) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
