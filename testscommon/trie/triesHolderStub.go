package trie

import (
	"github.com/multiversx/mx-chain-go/common"
)

// TriesHolderStub -
type TriesHolderStub struct {
	PutCalled         func([]byte, common.Trie)
	RemoveCalled      func([]byte, common.Trie)
	GetCalled         func([]byte) common.Trie
	GetAllDirtyCalled func() []common.Trie
	ResetCalled       func()
}

// Put -
func (ths *TriesHolderStub) Put(key []byte, trie common.Trie) {
	if ths.PutCalled != nil {
		ths.PutCalled(key, trie)
	}
}

// Get -
func (ths *TriesHolderStub) Get(key []byte) common.Trie {
	if ths.GetCalled != nil {
		return ths.GetCalled(key)
	}
	return nil
}

// GetAllDirty -
func (ths *TriesHolderStub) GetAllDirty() []common.Trie {
	if ths.GetAllDirtyCalled != nil {
		return ths.GetAllDirtyCalled()
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
