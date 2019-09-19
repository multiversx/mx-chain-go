package evictionWaitingList

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// evictionWaitingList is a structure that caches keys that need to be removed from a certain database.
// If the cache is full, the keys will be stored in the underlying database. Writing at the same key in
// cacher and db will overwrite the previous values
type evictionWaitingList struct {
	cache       map[string][][]byte
	cacheSize   int
	db          storage.Persister
	marshalizer marshal.Marshalizer
}

// NewEvictionWaitingList creates a new instance of evictionWaitingList
func NewEvictionWaitingList(size int, db storage.Persister, marshalizer marshal.Marshalizer) (*evictionWaitingList, error) {
	if size < 1 {
		return nil, data.ErrInvalidCacheSize
	}
	if db == nil || db.IsInterfaceNil() {
		return nil, data.ErrNilDatabase
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, data.ErrNilMarshalizer
	}

	return &evictionWaitingList{
		cache:       make(map[string][][]byte),
		cacheSize:   size,
		db:          db,
		marshalizer: marshalizer,
	}, nil
}

// Add adds the given hashes to the eviction waiting list, in the position given by the root hash
func (ec *evictionWaitingList) Add(rootHash []byte, hashes [][]byte) error {
	if len(ec.cache) < ec.cacheSize {
		ec.cache[string(rootHash)] = hashes
		return nil
	}

	marshalizedHashes, err := ec.marshalizer.Marshal(hashes)
	if err != nil {
		return err
	}

	err = ec.db.Put(rootHash, marshalizedHashes)
	if err != nil {
		return err
	}

	return nil
}

// Evict returns all the hashes from the position given by the root hash
func (ec *evictionWaitingList) Evict(rootHash []byte) ([][]byte, error) {
	hashes := ec.cache[string(rootHash)]
	if hashes != nil {
		delete(ec.cache, string(rootHash))
		return hashes, nil
	}

	marshalizedHashes, err := ec.db.Get(rootHash)
	if err != nil {
		return nil, err
	}

	err = ec.marshalizer.Unmarshal(&hashes, marshalizedHashes)
	if err != nil {
		return nil, err
	}

	err = ec.db.Remove(rootHash)
	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// Rollback clears the hashes from the position given by the root hash
func (ec *evictionWaitingList) Rollback(rootHash []byte) error {
	hashes := ec.cache[string(rootHash)]
	if hashes != nil {
		delete(ec.cache, string(rootHash))
		return nil
	}

	err := ec.db.Remove(rootHash)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *evictionWaitingList) IsInterfaceNil() bool {
	if ec == nil {
		return true
	}
	return false
}
