package trie

import (
	"github.com/multiversx/mx-chain-go/common"
)

// DataTrieTrackerStub -
type DataTrieTrackerStub struct {
	RetrieveValueCalled func(key []byte) ([]byte, uint32, error)
	SaveKeyValueCalled  func(key []byte, value []byte) error
	SetDataTrieCalled   func(tr common.Trie)
	DataTrieCalled      func() common.Trie
	SaveDirtyDataCalled func(trie common.Trie) (map[string][]byte, error)
}

// RetrieveValue -
func (dtts *DataTrieTrackerStub) RetrieveValue(key []byte) ([]byte, uint32, error) {
	if dtts.RetrieveValueCalled != nil {
		return dtts.RetrieveValueCalled(key)
	}

	return []byte{}, 0, nil
}

// SaveKeyValue -
func (dtts *DataTrieTrackerStub) SaveKeyValue(key []byte, value []byte) error {
	if dtts.SaveKeyValueCalled != nil {
		return dtts.SaveKeyValueCalled(key, value)
	}

	return nil
}

// SetDataTrie -
func (dtts *DataTrieTrackerStub) SetDataTrie(tr common.Trie) {
	if dtts.SetDataTrieCalled != nil {
		dtts.SetDataTrieCalled(tr)
	}
}

// DataTrie -
func (dtts *DataTrieTrackerStub) DataTrie() common.DataTrieHandler {
	if dtts.DataTrieCalled != nil {
		return dtts.DataTrieCalled()
	}

	return nil
}

// SaveDirtyData -
func (dtts *DataTrieTrackerStub) SaveDirtyData(mainTrie common.Trie) (map[string][]byte, error) {
	if dtts.SaveDirtyDataCalled != nil {
		return dtts.SaveDirtyDataCalled(mainTrie)
	}

	return map[string][]byte{}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dtts *DataTrieTrackerStub) IsInterfaceNil() bool {
	return dtts == nil
}
