package mock

import "github.com/ElrondNetwork/elrond-go/data"

// DataTrieTrackerStub -
type DataTrieTrackerStub struct {
	ClearDataCachesCalled func()
	DirtyDataCalled       func() map[string][]byte
	RetrieveValueCalled   func(key []byte) ([]byte, error)
	SaveKeyValueCalled    func(key []byte, value []byte) error
	SetDataTrieCalled     func(tr data.Trie)
	DataTrieCalled        func() data.Trie
}

// ClearDataCaches -
func (dtts *DataTrieTrackerStub) ClearDataCaches() {
	if dtts.ClearDataCachesCalled != nil {
		dtts.ClearDataCachesCalled()
	}
}

// DirtyData -
func (dtts *DataTrieTrackerStub) DirtyData() map[string][]byte {
	if dtts.DirtyDataCalled != nil {
		return dtts.DirtyDataCalled()
	}
	return nil
}

// RetrieveValue -
func (dtts *DataTrieTrackerStub) RetrieveValue(key []byte) ([]byte, error) {
	if dtts.RetrieveValueCalled != nil {
		return dtts.RetrieveValueCalled(key)
	}
	return nil, nil
}

// SaveKeyValue -
func (dtts *DataTrieTrackerStub) SaveKeyValue(key []byte, value []byte) error {
	if dtts.SaveKeyValueCalled != nil {
		return dtts.SaveKeyValueCalled(key, value)
	}
	return nil
}

// SetDataTrie -
func (dtts *DataTrieTrackerStub) SetDataTrie(tr data.Trie) {
	if dtts.SetDataTrieCalled != nil {
		dtts.SetDataTrieCalled(tr)
	}
}

// DataTrie -
func (dtts *DataTrieTrackerStub) DataTrie() data.Trie {
	if dtts.DataTrieCalled != nil {
		return dtts.DataTrieCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dtts *DataTrieTrackerStub) IsInterfaceNil() bool {
	return dtts == nil
}
