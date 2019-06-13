package leveldb

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
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
	dbClosed          chan struct{}
	dbAccess          chan serialQueryer
	ctx               context.Context
}

// NewSerialDB is a constructor for the leveldb persister
// It creates the files in the location given as parameter
func NewSerialDB(path string, batchDelaySeconds int, maxBatchSize int) (s *SerialDB, err error) {
	err = os.MkdirAll(path, rwxOwner)
	if err != nil {
		return nil, err
	}

	options := &opt.Options{
		// disable internal cache
		BlockCacheCapacity: -1,
	}

	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	dbStore := &SerialDB{
		db:                db,
		path:              path,
		maxBatchSize:      maxBatchSize,
		batchDelaySeconds: batchDelaySeconds,
		sizeBatch:         0,
		dbClosed:          make(chan struct{}),
		dbAccess:          make(chan serialQueryer),
		ctx:               ctx,
	}

	dbStore.batch = dbStore.createBatch()

	go dbStore.batchTimeoutHandle()
	go dbStore.processLoop(ctx)

	return dbStore, nil
}

func (s *SerialDB) batchTimeoutHandle() {
	for {
		select {
		case <-time.After(time.Duration(s.batchDelaySeconds) * time.Second):
			err := s.putBatch()
			if err != nil {
				log.Error(err.Error())
				continue
			}
		case <-s.dbClosed:
			return
		}
	}
}

// Put adds the value to the (key, val) storage medium
func (s *SerialDB) Put(key, val []byte) error {
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
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

// Get returns the value associated to the key
func (s *SerialDB) Get(key []byte) ([]byte, error) {
	ch := make(chan *pairResult)
	req := &getAct{
		key:     key,
		resChan: ch,
	}

	s.dbAccess <- req
	p := <-ch
	close(ch)

	if p.err == leveldb.ErrNotFound {
		return nil, storage.ErrKeyNotFound
	}

	return p.value, nil
}

// Has returns true if the given key is present in the persistence medium
func (s *SerialDB) Has(key []byte) error {
	ch := make(chan error)
	req := &hasAct{
		key:     key,
		resChan: ch,
	}

	s.dbAccess <- req
	p := <-ch
	close(ch)

	return p
}

// Init initializes the storage medium and prepares it for usage
func (s *SerialDB) Init() error {
	// no special initialization needed
	return nil
}

// CreateBatch returns a batcher to be used for batch writing data to the database
func (s *SerialDB) createBatch() storage.Batcher {
	return NewBatch()
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
	s.batch = s.createBatch()
	s.mutBatch.Unlock()

	ch := make(chan error)
	//request
	req := &putBatchAct{
		batch:   batch,
		resChan: ch,
	}

	s.dbAccess <- req
	// await response
	p := <-ch
	close(ch)

	return p
}

// Close closes the files/resources associated to the storage medium
func (s *SerialDB) Close() error {
	_ = s.putBatch()

	_, cancel := context.WithCancel(s.ctx)
	cancel()
	s.dbClosed <- struct{}{}

	return s.db.Close()
}

// Remove removes the data associated to the given key
func (s *SerialDB) Remove(key []byte) error {
	s.mutBatch.Lock()
	_ = s.batch.Delete(key)
	s.mutBatch.Unlock()

	ch := make(chan error)
	//request
	req := &delAct{
		key:     key,
		resChan: ch,
	}

	s.dbAccess <- req
	// await response
	p := <-ch
	close(ch)

	return p
}

// Destroy removes the storage medium stored data
func (s *SerialDB) Destroy() error {
	s.mutBatch.Lock()
	s.batch.Reset()
	s.sizeBatch = 0
	s.mutBatch.Unlock()

	_, cancel := context.WithCancel(s.ctx)
	cancel()
	s.dbClosed <- struct{}{}
	err := s.db.Close()
	if err != nil {
		return err
	}

	err = os.RemoveAll(s.path)

	return err
}

func (s *SerialDB) processLoop(ctx context.Context) {
	defer func() {
		// TODO: clean up if needed
	}()

	for {
		select {
		case queryer := <-s.dbAccess:
			queryer.request(s)
		case <-ctx.Done():
			log.Info("closing the leveldb process loop")
			return
		}
	}
}
