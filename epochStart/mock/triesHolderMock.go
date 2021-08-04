package mock

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

// TriesHolderMock -
type TriesHolderMock struct {
	PutCalled    func([]byte, common.Trie)
	RemoveCalled func([]byte, common.Trie)
	GetCalled    func([]byte) common.Trie
	GetAllCalled func() []common.Trie
	ResetCalled  func()
}

// Put -
func (thm *TriesHolderMock) Put(key []byte, trie common.Trie) {
	if thm.PutCalled != nil {
		thm.PutCalled(key, trie)
	}
}

// Replace -
func (thm *TriesHolderMock) Replace(key []byte, trie common.Trie) {
	if thm.RemoveCalled != nil {
		thm.RemoveCalled(key, trie)
	}
}

// Get -
func (thm *TriesHolderMock) Get(key []byte) common.Trie {
	if thm.GetCalled != nil {
		return thm.GetCalled(key)
	}
	return nil
}

// GetAll -
func (thm *TriesHolderMock) GetAll() []common.Trie {
	if thm.GetAllCalled != nil {
		return thm.GetAllCalled()
	}
	return nil
}

// Reset -
func (thm *TriesHolderMock) Reset() {
	if thm.ResetCalled != nil {
		thm.ResetCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (thm *TriesHolderMock) IsInterfaceNil() bool {
	return thm == nil
}
