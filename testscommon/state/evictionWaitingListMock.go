package state

import (
	"bytes"
	"errors"
	"sync"

	"github.com/ElrondNetwork/elrond-go/state"
)

// EvictionWaitingListMock is a structure that caches keys that need to be removed from a certain database.
// If the cache is full, the keys will be stored in the underlying database. Writing at the same key in
// cacher and db will overwrite the previous values.
type EvictionWaitingListMock struct {
	Cache     map[string][][]byte
	CacheSize uint
	OpMutex   sync.RWMutex
}

// NewEvictionWaitingListMock creates a new instance of evictionWaitingList
func NewEvictionWaitingListMock(size uint) *EvictionWaitingListMock {
	return &EvictionWaitingListMock{
		Cache:     make(map[string][][]byte),
		CacheSize: size,
	}
}

// Put stores the given hashes in the eviction waiting list, in the position given by the root hash
func (ewl *EvictionWaitingListMock) Put(rootHash []byte, hashes [][]byte) error {
	ewl.OpMutex.Lock()
	defer ewl.OpMutex.Unlock()

	if uint(len(ewl.Cache)) < ewl.CacheSize {
		ewl.Cache[string(rootHash)] = hashes
		return nil
	}

	return nil
}

// Evict returns and removes from the waiting list all the hashes from the position given by the root hash
func (ewl *EvictionWaitingListMock) Evict(rootHash []byte) ([][]byte, error) {
	ewl.OpMutex.Lock()
	defer ewl.OpMutex.Unlock()

	hashes, ok := ewl.Cache[string(rootHash)]
	if ok {
		delete(ewl.Cache, string(rootHash))
	}

	return hashes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ewl *EvictionWaitingListMock) IsInterfaceNil() bool {
	return ewl == nil
}

// ShouldKeepHash -
func (ewl *EvictionWaitingListMock) ShouldKeepHash(hash string, identifier state.TriePruningIdentifier) (bool, error) {
	ewl.OpMutex.Lock()
	defer ewl.OpMutex.Unlock()

	for key := range ewl.Cache {
		if len(key) == 0 {
			return false, errors.New("invalid key")
		}

		lastByte := key[len(key)-1]
		if state.TriePruningIdentifier(lastByte) == state.OldRoot && identifier == state.OldRoot {
			continue
		}

		hashes := ewl.Cache[key]

		for _, h := range hashes {
			if bytes.Equal(h, []byte(hash)) {
				return true, nil
			}
		}
	}

	return false, nil
}

// Close -
func (ewl *EvictionWaitingListMock) Close() error {
	return nil
}
