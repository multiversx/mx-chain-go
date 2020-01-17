package mock

import "github.com/ElrondNetwork/elrond-go/data"

type TriesHolderMock struct {
	PutCalled    func([]byte, data.Trie)
	GetCalled    func([]byte) data.Trie
	GetAllCalled func() []data.Trie
	ResetCalled  func()
}

func (thm *TriesHolderMock) Put(key []byte, trie data.Trie) {
	if thm.PutCalled != nil {
		thm.PutCalled(key, trie)
	}
}
func (thm *TriesHolderMock) Get(key []byte) data.Trie {
	if thm.GetCalled != nil {
		return thm.GetCalled(key)
	}
	return nil
}
func (thm *TriesHolderMock) GetAll() []data.Trie {
	if thm.GetAllCalled != nil {
		return thm.GetAllCalled()
	}
	return nil
}
func (thm *TriesHolderMock) Reset() {
	if thm.ResetCalled != nil {
		thm.ResetCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (thm *TriesHolderMock) IsInterfaceNil() bool {
	return thm == nil
}
