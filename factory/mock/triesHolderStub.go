package mock

import (
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

// TriesHolderStub -
type TriesHolderStub struct {
	PutCalled    func([]byte, temporary.Trie)
	RemoveCalled func([]byte, temporary.Trie)
	GetCalled    func([]byte) temporary.Trie
	GetAllCalled func() []temporary.Trie
	ResetCalled  func()
}

// Put -
func (ths *TriesHolderStub) Put(key []byte, trie temporary.Trie) {
	if ths.PutCalled != nil {
		ths.PutCalled(key, trie)
	}
}

// Replace -
func (ths *TriesHolderStub) Replace(key []byte, trie temporary.Trie) {
	if ths.RemoveCalled != nil {
		ths.RemoveCalled(key, trie)
	}
}

// Get -
func (ths *TriesHolderStub) Get(key []byte) temporary.Trie {
	if ths.GetCalled != nil {
		return ths.GetCalled(key)
	}
	return &testscommon.TrieStub{}
}

// GetAll -
func (ths *TriesHolderStub) GetAll() []temporary.Trie {
	if ths.GetAllCalled != nil {
		return ths.GetAllCalled()
	}
	return nil
}

// Reset -
func (ths *TriesHolderStub) Reset() {
	if ths.ResetCalled != nil {
		ths.ResetCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ths *TriesHolderStub) IsInterfaceNil() bool {
	return ths == nil
}
