package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/trie"
)

type TrieStub struct {
	SetCacheLimitCalled func(l uint16)
	GetCalled           func(key []byte) ([]byte, error)
	UpdateCalled        func(key, value []byte) error
	DeleteCalled        func(key []byte) error
	RootCalled          func() []byte
	CommitCalled        func(onleaf trie.LeafCallback) (root []byte, err error)
	DBWCalled           func() trie.DBWriteCacher
	RecreateCalled      func(root []byte, dbw trie.DBWriteCacher) (trie.PatriciaMerkelTree, error)
	CopyCalled          func() trie.PatriciaMerkelTree
}

func (ts *TrieStub) SetCacheLimit(l uint16) {
	ts.SetCacheLimitCalled(l)
}

func (ts *TrieStub) Get(key []byte) ([]byte, error) {
	return ts.GetCalled(key)
}

func (ts *TrieStub) Update(key, value []byte) error {
	return ts.UpdateCalled(key, value)
}

func (ts *TrieStub) Delete(key []byte) error {
	return ts.DeleteCalled(key)
}

func (ts *TrieStub) Root() []byte {
	return ts.RootCalled()
}

func (ts *TrieStub) Commit(onleaf trie.LeafCallback) (root []byte, err error) {
	return ts.CommitCalled(onleaf)
}

func (ts *TrieStub) DBW() trie.DBWriteCacher {
	return ts.DBWCalled()
}

func (ts *TrieStub) Recreate(root []byte, dbw trie.DBWriteCacher) (trie.PatriciaMerkelTree, error) {
	return ts.RecreateCalled(root, dbw)
}

func (ts *TrieStub) Copy() trie.PatriciaMerkelTree {
	return ts.CopyCalled()
}
