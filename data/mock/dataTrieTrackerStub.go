package mock

import "github.com/ElrondNetwork/elrond-go/data/trie"

type DataTrieTrackerStub struct {
	ClearDataCachesCalled func()
	DirtyDataCalled       func() map[string][]byte
	OriginalValueCalled   func(key []byte) []byte
	RetrieveValueCalled   func(key []byte) ([]byte, error)
	SaveKeyValueCalled    func(key []byte, value []byte)
	SetDataTrieCalled     func(tr trie.PatriciaMerkelTree)
	DataTrieCalled        func() trie.PatriciaMerkelTree
}

func (dtts *DataTrieTrackerStub) ClearDataCaches() {
	dtts.ClearDataCachesCalled()
}

func (dtts *DataTrieTrackerStub) DirtyData() map[string][]byte {
	return dtts.DirtyDataCalled()
}

func (dtts *DataTrieTrackerStub) OriginalValue(key []byte) []byte {
	return dtts.OriginalValueCalled(key)
}

func (dtts *DataTrieTrackerStub) RetrieveValue(key []byte) ([]byte, error) {
	return dtts.RetrieveValueCalled(key)
}

func (dtts *DataTrieTrackerStub) SaveKeyValue(key []byte, value []byte) {
	dtts.SaveKeyValueCalled(key, value)
}

func (dtts *DataTrieTrackerStub) SetDataTrie(tr trie.PatriciaMerkelTree) {
	dtts.SetDataTrieCalled(tr)
}

func (dtts *DataTrieTrackerStub) DataTrie() trie.PatriciaMerkelTree {
	return dtts.DataTrieCalled()
}
