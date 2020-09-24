package evictionWaitingList

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("trie")

// evictionWaitingList is a structure that caches keys that need to be removed from a certain database.
// If the cache is full, the keys will be stored in the underlying database. Writing at the same key in
// cacher and db will overwrite the previous values.
type evictionWaitingList struct {
	cache       map[string]data.ModifiedHashes
	cacheSize   uint
	db          storage.Persister
	marshalizer marshal.Marshalizer
	opMutex     sync.RWMutex
}

// NewEvictionWaitingList creates a new instance of evictionWaitingList
func NewEvictionWaitingList(size uint, db storage.Persister, marshalizer marshal.Marshalizer) (*evictionWaitingList, error) {
	if size < 1 {
		return nil, data.ErrInvalidCacheSize
	}
	if check.IfNil(db) {
		return nil, data.ErrNilDatabase
	}
	if check.IfNil(marshalizer) {
		return nil, data.ErrNilMarshalizer
	}

	return &evictionWaitingList{
		cache:       make(map[string]data.ModifiedHashes),
		cacheSize:   size,
		db:          db,
		marshalizer: marshalizer,
	}, nil
}

// Put stores the given hashes in the eviction waiting list, in the position given by the root hash
func (ewl *evictionWaitingList) Put(rootHash []byte, hashes data.ModifiedHashes) error {
	ewl.opMutex.Lock()
	defer ewl.opMutex.Unlock()

	log.Trace("trie eviction waiting list", "size", len(ewl.cache))

	if uint(len(ewl.cache)) < ewl.cacheSize {
		ewl.cache[string(rootHash)] = hashes
		return nil
	}

	b := &batch.Batch{}

	for h := range hashes {
		b.Data = append(b.Data, []byte(h))
	}

	marshalizedHashes, err := ewl.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	err = ewl.db.Put(rootHash, marshalizedHashes)
	if err != nil {
		return err
	}
	ewl.cache[string(rootHash)] = nil

	return nil
}

// Evict returns and removes from the waiting list all the hashes from the position given by the root hash
func (ewl *evictionWaitingList) Evict(rootHash []byte) (data.ModifiedHashes, error) {
	ewl.opMutex.Lock()
	defer ewl.opMutex.Unlock()

	hashes, ok := ewl.cache[string(rootHash)]

	if !ok {
		return nil, nil
	}

	delete(ewl.cache, string(rootHash))
	if len(hashes) != 0 {
		return hashes, nil
	}

	marshalizedHashes, err := ewl.db.Get(rootHash)
	if err != nil {
		return nil, err
	}

	b := &batch.Batch{}

	err = ewl.marshalizer.Unmarshal(b, marshalizedHashes)
	if err != nil {
		return nil, err
	}

	hashes = make(data.ModifiedHashes, len(b.Data))
	for _, h := range b.Data {
		hashes[string(h)] = struct{}{}
	}

	err = ewl.db.Remove(rootHash)
	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ewl *evictionWaitingList) IsInterfaceNil() bool {
	return ewl == nil
}

// ShouldKeepHash searches for the given hash in all of the evictionWaitingList's newHashes.
// If the identifier is equal to oldRoot, then we should also search in oldHashes.
func (ewl *evictionWaitingList) ShouldKeepHash(hash string, identifier data.TriePruningIdentifier) (bool, error) {
	ewl.opMutex.RLock()
	defer ewl.opMutex.RUnlock()

	for key := range ewl.cache {
		if len(key) == 0 {
			return false, ErrInvalidKey
		}

		lastByte := key[len(key)-1]
		if data.TriePruningIdentifier(lastByte) == data.OldRoot && identifier == data.OldRoot {
			continue
		}

		hashes := ewl.cache[key]
		if len(hashes) == 0 {
			marshalizedHashes, err := ewl.db.Get([]byte(key))
			if err != nil {
				return false, err
			}

			b := &batch.Batch{}

			err = ewl.marshalizer.Unmarshal(b, marshalizedHashes)
			if err != nil {
				return false, err
			}

			hashes = make(data.ModifiedHashes, len(b.Data))
			for _, h := range b.Data {
				hashes[string(h)] = struct{}{}
			}
		}
		_, ok := hashes[hash]
		if ok {
			log.Trace("should keep hash", "rootHash", []byte(key), "hash", hash)
			return true, nil
		}
	}

	return false, nil
}

// Close - closes the underlying db
func (ewl *evictionWaitingList) Close() error {
	if !check.IfNil(ewl.db) {
		//TODO: @beni verify if a flush is needed before closing the DB
		return ewl.db.Close()
	}
	return nil
}
