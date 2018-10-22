package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// ConcurrentTrie is a trie implementation that sould be concurrent safe
type ConcurrentTrie struct {
	*trie.Trie
	mutTrie sync.RWMutex
}

// NewConcurrentTrie returns a new trie that should be concurrent safe
func NewConcurrentTrie(root []byte, dbw trie.DBWriteCacher, hsh hashing.Hasher) (*ConcurrentTrie, error) {
	tr, err := trie.NewTrie(root, dbw, hsh)

	if err != nil {
		return nil, err
	}

	ct := ConcurrentTrie{Trie: tr, mutTrie: sync.RWMutex{}}

	return &ct, nil
}

// SetCacheLimit calls the underlying trie's equivalent method
func (ct *ConcurrentTrie) SetCacheLimit(l uint16) {
	ct.Trie.SetCacheLimit(l)
}

// Get fetches data from the underlying trie. Concurrent safe
func (ct *ConcurrentTrie) Get(key []byte) ([]byte, error) {
	ct.mutTrie.RLock()
	defer ct.mutTrie.RUnlock()

	return ct.Trie.Get(key)
}

// Update changes data from the underlying trie. Concurrent safe
func (ct *ConcurrentTrie) Update(key, value []byte) error {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Update(key, value)
}

// Deletes data from the underlying trie. Concurrent safe
func (ct *ConcurrentTrie) Delete(key []byte) error {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Delete(key)
}

// Root computes trie root hash of the underlying trie. Concurrent safe
func (ct *ConcurrentTrie) Root() []byte {
	ct.mutTrie.RLock()
	defer ct.mutTrie.RUnlock()

	return ct.Trie.Root()
}

// Commit calls underlying trie's commit method. Concurrent safe
func (ct *ConcurrentTrie) Commit(onleaf trie.LeafCallback) (root []byte, err error) {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Commit(onleaf)
}

// DBW returns the underlying trie's DatabaseWriteCacher object
func (ct *ConcurrentTrie) DBW() trie.DBWriteCacher {
	return ct.Trie.DBW()
}

// Recreates the trie from the input hash. Concurrent safe
func (ct *ConcurrentTrie) Recreate(root []byte, dbw trie.DBWriteCacher) (trie.PatriciaMerkelTree, error) {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Recreate(root, dbw)
}

// Copy will call underlying trie's Copy() func. Concurrent safe
func (ct *ConcurrentTrie) Copy() trie.PatriciaMerkelTree {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Copy()
}
