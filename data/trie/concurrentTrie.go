package trie

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"sync"
)

//A trie implementation that sould be concurrent safe

type ConcurrentTrie struct {
	*Trie
	mutTrie sync.RWMutex
}

func NewConcurrentTrie(root []byte, dbw DBWriteCacher, hsh hashing.Hasher) (*ConcurrentTrie, error) {
	tr, err := NewTrie(root, dbw, hsh)

	if err != nil {
		return nil, err
	}

	ct := ConcurrentTrie{Trie: tr, mutTrie: sync.RWMutex{}}

	return &ct, nil
}

func (ct *ConcurrentTrie) SetCacheLimit(l uint16) {
	ct.Trie.SetCacheLimit(l)
}

func (ct *ConcurrentTrie) Get(key []byte) ([]byte, error) {
	ct.mutTrie.RLock()
	defer ct.mutTrie.RUnlock()

	return ct.Trie.Get(key)
}

func (ct *ConcurrentTrie) Update(key, value []byte) error {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Update(key, value)
}

func (ct *ConcurrentTrie) Delete(key []byte) error {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Delete(key)
}

func (ct *ConcurrentTrie) Root() []byte {
	ct.mutTrie.RLock()
	defer ct.mutTrie.RUnlock()

	return ct.Trie.Root()
}

func (ct *ConcurrentTrie) Commit(onleaf LeafCallback) (root []byte, err error) {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Commit(onleaf)
}

func (ct *ConcurrentTrie) DBW() DBWriteCacher {
	return ct.DBW()
}

func (ct *ConcurrentTrie) Recreate(root []byte, dbw DBWriteCacher) (Trier, error) {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Recreate(root, dbw)
}

func (ct *ConcurrentTrie) Copy() Trier {
	ct.mutTrie.Lock()
	defer ct.mutTrie.Unlock()

	return ct.Trie.Copy()
}
