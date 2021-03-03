package mock

import "github.com/ElrondNetwork/elrond-go/data"

// TriesHolderStub -
type TriesHolderStub struct {
	PutCalled    func([]byte, data.Trie)
	RemoveCalled func([]byte, data.Trie)
	GetCalled    func([]byte) data.Trie
	GetAllCalled func() []data.Trie
	ResetCalled  func()
}

// Put -
func (ths *TriesHolderStub) Put(key []byte, trie data.Trie) {
	if ths.PutCalled != nil {
		ths.PutCalled(key, trie)
	}
}

// Replace -
func (ths *TriesHolderStub) Replace(key []byte, trie data.Trie) {
	if ths.RemoveCalled != nil {
		ths.RemoveCalled(key, trie)
	}
}

// Get -
func (ths *TriesHolderStub) Get(key []byte) data.Trie {
	if ths.GetCalled != nil {
		return ths.GetCalled(key)
	}
	return nil
}

// GetAll -
func (ths *TriesHolderStub) GetAll() []data.Trie {
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
