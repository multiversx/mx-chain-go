package trie

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// DataTrieTrackerStub -
type DataTrieTrackerStub struct {
	dataTrie common.Trie

	RetrieveValueCalled         func(key []byte) ([]byte, uint32, error)
	SaveKeyValueCalled          func(key []byte, value []byte) error
	SetDataTrieCalled           func(tr common.Trie)
	DataTrieCalled              func() common.Trie
	SaveDirtyDataCalled         func(trie common.Trie) ([]core.TrieData, error)
	SaveTrieDataCalled          func(trieData core.TrieData) error
	MigrateDataTrieLeavesCalled func(args vmcommon.ArgsMigrateDataTrieLeaves) error
}

// RetrieveValue -
func (dtts *DataTrieTrackerStub) RetrieveValue(key []byte) ([]byte, uint32, error) {
	if dtts.RetrieveValueCalled != nil {
		return dtts.RetrieveValueCalled(key)
	}

	return []byte{0}, 0, nil
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

	dtts.dataTrie = tr
}

// DataTrie -
func (dtts *DataTrieTrackerStub) DataTrie() common.DataTrieHandler {
	if dtts.DataTrieCalled != nil {
		return dtts.DataTrieCalled()
	}

	if !check.IfNil(dtts.dataTrie) {
		return dtts.dataTrie
	}

	return nil
}

// SaveDirtyData -
func (dtts *DataTrieTrackerStub) SaveDirtyData(mainTrie common.Trie) ([]core.TrieData, error) {
	if dtts.SaveDirtyDataCalled != nil {
		return dtts.SaveDirtyDataCalled(mainTrie)
	}

	return make([]core.TrieData, 0), nil
}

// MigrateDataTrieLeaves -
func (dtts *DataTrieTrackerStub) MigrateDataTrieLeaves(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	if dtts.MigrateDataTrieLeavesCalled != nil {
		return dtts.MigrateDataTrieLeavesCalled(args)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dtts *DataTrieTrackerStub) IsInterfaceNil() bool {
	return dtts == nil
}
