package leveldb

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var _ storage.Persister = (*DB)(nil)

// read + write + execute for owner only
const rwxOwner = 0700

var log = logger.GetOrCreate("storage/leveldb")

// DB holds a pointer to the leveldb database and the path to where it is stored.
type DB struct {
	*baseLevelDb
	path              string
	maxBatchSize      int
	batchDelaySeconds int
	sizeBatch         int
	batch             storage.Batcher
	mutBatch          sync.RWMutex
	dbClosed          chan struct{}
}

// NewDB is a constructor for the leveldb persister
// It creates the files in the location given as parameter
func NewDB(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (s *DB, err error) {
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

	dbStore := &DB{
		baseLevelDb:       bldb,
		path:              path,
		maxBatchSize:      maxBatchSize,
		batchDelaySeconds: batchDelaySeconds,
		sizeBatch:         0,
		dbClosed:          make(chan struct{}),
	}

	dbStore.batch = dbStore.createBatch()

	go dbStore.batchTimeoutHandle()

	runtime.SetFinalizer(dbStore, func(db *DB) {
		_ = db.Close()
	})

	log.Debug("opened level db persister", "path", path)

	return dbStore, nil
}

func (s *DB) batchTimeoutHandle() {
	for {
		select {
		case <-time.After(time.Duration(s.batchDelaySeconds) * time.Second):
			s.mutBatch.Lock()
			err := s.putBatch(s.batch)
			if err != nil {
				log.Warn("leveldb putBatch", "error", err.Error())
				s.mutBatch.Unlock()
				continue
			}

			s.batch.Reset()
			s.sizeBatch = 0
			s.mutBatch.Unlock()
		case <-s.dbClosed:
			log.Debug("closing the timed batch handler", "path", s.path)
			return
		}
	}
}

func (s *DB) updateBatchWithIncrement() error {
	s.mutBatch.Lock()
	defer s.mutBatch.Unlock()

	s.sizeBatch++
	if s.sizeBatch < s.maxBatchSize {
		return nil
	}

	err := s.putBatch(s.batch)
	if err != nil {
		log.Warn("leveldb putBatch", "error", err.Error())
		return err
	}

	s.batch.Reset()
	s.sizeBatch = 0

	return nil
}

// Put adds the value to the (key, val) storage medium
func (s *DB) Put(key, val []byte) error {
	err := s.batch.Put(key, val)
	if err != nil {
		return err
	}

	return s.updateBatchWithIncrement()
}

// Get returns the value associated to the key
func (s *DB) Get(key []byte) ([]byte, error) {
	data := s.batch.Get(key)
	if data != nil {
		if bytes.Equal(data, []byte(removed)) {
			return nil, storage.ErrKeyNotFound
		}
		return data, nil
	}

	data, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, storage.ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Has returns nil if the given key is present in the persistence medium
func (s *DB) Has(key []byte) error {
	data := s.batch.Get(key)
	if data != nil {
		if bytes.Equal(data, []byte(removed)) {
			return storage.ErrKeyNotFound
		}
		return nil
	}

	has, err := s.db.Has(key, nil)
	if err != nil {
		return err
	}

	if has {
		return nil
	}

	return storage.ErrKeyNotFound
}

// Init initializes the storage medium and prepares it for usage
func (s *DB) Init() error {
	// no special initialization needed
	return nil
}

// CreateBatch returns a batcher to be used for batch writing data to the database
func (s *DB) createBatch() storage.Batcher {
	return NewBatch()
}

// putBatch writes the Batch data into the database
func (s *DB) putBatch(b storage.Batcher) error {
	dbBatch, ok := b.(*batch)
	if !ok {
		return storage.ErrInvalidBatch
	}

	wopt := &opt.WriteOptions{
		Sync: true,
	}

	return s.db.Write(dbBatch.batch, wopt)
}

// Close closes the files/resources associated to the storage medium
func (s *DB) Close() error {
	s.mutBatch.Lock()
	_ = s.putBatch(s.batch)
	s.sizeBatch = 0
	s.mutBatch.Unlock()

	select {
	case s.dbClosed <- struct{}{}:
	default:
	}

	return s.db.Close()
}

// Remove removes the data associated to the given key
func (s *DB) Remove(key []byte) error {
	s.mutBatch.Lock()
	_ = s.batch.Delete(key)
	s.mutBatch.Unlock()

	return s.updateBatchWithIncrement()
}

// Destroy removes the storage medium stored data
func (s *DB) Destroy() error {
	s.mutBatch.Lock()
	s.batch.Reset()
	s.sizeBatch = 0
	s.mutBatch.Unlock()

	s.dbClosed <- struct{}{}
	err := s.db.Close()
	if err != nil {
		return err
	}

	err = os.RemoveAll(s.path)

	return err
}

// DestroyClosed removes the already closed storage medium stored data
func (s *DB) DestroyClosed() error {
	return os.RemoveAll(s.path)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *DB) IsInterfaceNil() bool {
	return s == nil
}
