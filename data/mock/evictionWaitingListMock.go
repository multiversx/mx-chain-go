package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// EvictionWaitingList is a structure that caches keys that need to be removed from a certain database.
// If the cache is full, the keys will be stored in the underlying database. Writing at the same key in
// cacher and db will overwrite the previous values. This structure is not concurrent safe.
type EvictionWaitingList struct {
	Cache       map[string][][]byte
	CacheSize   int
	Db          storage.Persister
	Marshalizer marshal.Marshalizer
}

// NewEvictionWaitingList creates a new instance of evictionWaitingList
func NewEvictionWaitingList(size int, db storage.Persister, marshalizer marshal.Marshalizer) (*EvictionWaitingList, error) {
	if size < 1 {
		return nil, data.ErrInvalidCacheSize
	}
	if db == nil || db.IsInterfaceNil() {
		return nil, data.ErrNilDatabase
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, data.ErrNilMarshalizer
	}

	return &EvictionWaitingList{
		Cache:       make(map[string][][]byte),
		CacheSize:   size,
		Db:          db,
		Marshalizer: marshalizer,
	}, nil
}

// Put stores the given hashes in the eviction waiting list, in the position given by the root hash
func (ewl *EvictionWaitingList) Put(rootHash []byte, hashes [][]byte) error {
	if len(ewl.Cache) < ewl.CacheSize {
		ewl.Cache[string(rootHash)] = hashes
		return nil
	}

	marshalizedHashes, err := ewl.Marshalizer.Marshal(hashes)
	if err != nil {
		return err
	}

	err = ewl.Db.Put(rootHash, marshalizedHashes)
	if err != nil {
		return err
	}

	return nil
}

// Evict returns and removes from the waiting list all the hashes from the position given by the root hash
func (ewl *EvictionWaitingList) Evict(rootHash []byte) ([][]byte, error) {
	hashes, ok := ewl.Cache[string(rootHash)]
	if ok {
		delete(ewl.Cache, string(rootHash))
		return hashes, nil
	}

	marshalizedHashes, err := ewl.Db.Get(rootHash)
	if err != nil {
		return nil, err
	}

	err = ewl.Marshalizer.Unmarshal(&hashes, marshalizedHashes)
	if err != nil {
		return nil, err
	}

	err = ewl.Db.Remove(rootHash)
	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ewl *EvictionWaitingList) IsInterfaceNil() bool {
	if ewl == nil {
		return true
	}
	return false
}

// GetSize returns the size of the cache
func (ewl *EvictionWaitingList) GetSize() int {
	return ewl.CacheSize
}
