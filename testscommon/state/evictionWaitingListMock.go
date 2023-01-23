package state

import (
	"errors"
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// EvictionWaitingListMock is a mock implementation of state.DBRemoveCacher
type EvictionWaitingListMock struct {
	Cache     map[string]common.ModifiedHashes
	CacheSize uint
	OpMutex   sync.RWMutex
}

// NewEvictionWaitingListMock creates a new instance of evictionWaitingList
func NewEvictionWaitingListMock(size uint) *EvictionWaitingListMock {
	return &EvictionWaitingListMock{
		Cache:     make(map[string]common.ModifiedHashes),
		CacheSize: size,
	}
}

// Put stores the given hashes in the eviction waiting list, in the position given by the root hash
func (ewl *EvictionWaitingListMock) Put(rootHash []byte, hashes common.ModifiedHashes) error {
	ewl.OpMutex.Lock()
	defer ewl.OpMutex.Unlock()

	if uint(len(ewl.Cache)) < ewl.CacheSize {
		ewl.Cache[string(rootHash)] = hashes
		return nil
	}

	return nil
}

// Evict returns and removes from the waiting list all the hashes from the position given by the root hash
func (ewl *EvictionWaitingListMock) Evict(rootHash []byte) (common.ModifiedHashes, error) {
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

		for h := range hashes {
			if h == hash {
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
