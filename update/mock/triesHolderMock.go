package mock

import (
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// TriesHolderMock -
type TriesHolderMock struct {
	PutCalled    func([]byte, temporary.Trie)
	RemoveCalled func([]byte, temporary.Trie)
	GetCalled    func([]byte) temporary.Trie
	GetAllCalled func() []temporary.Trie
	ResetCalled  func()
}

// Put -
func (thm *TriesHolderMock) Put(key []byte, trie temporary.Trie) {
	if thm.PutCalled != nil {
		thm.PutCalled(key, trie)
	}
}

// Replace -
func (thm *TriesHolderMock) Replace(key []byte, trie temporary.Trie) {
	if thm.RemoveCalled != nil {
		thm.RemoveCalled(key, trie)
	}
}

// Get -
func (thm *TriesHolderMock) Get(key []byte) temporary.Trie {
	if thm.GetCalled != nil {
		return thm.GetCalled(key)
	}
	return nil
}

// GetAll -
func (thm *TriesHolderMock) GetAll() []temporary.Trie {
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
