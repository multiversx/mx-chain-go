package disabled

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type disabledTrackableDataTrie struct {
}

// NewDisabledTrackableDataTrie returns a new instance of disabledTrackableDataTrie
func NewDisabledTrackableDataTrie() *disabledTrackableDataTrie {
	return &disabledTrackableDataTrie{}
}

// RetrieveValue returns an empty byte array
func (dtdt *disabledTrackableDataTrie) RetrieveValue(_ []byte) ([]byte, uint32, error) {
	return []byte{}, 0, nil
}

// SaveKeyValue does nothing for this implementation
func (dtdt *disabledTrackableDataTrie) SaveKeyValue(_ []byte, _ []byte) error {
	return nil
}

// SetDataTrie does nothing for this implementation
func (dtdt *disabledTrackableDataTrie) SetDataTrie(_ common.Trie) {
}

// DataTrie returns a new disabledDataTrieHandler
func (dtdt *disabledTrackableDataTrie) DataTrie() common.DataTrieHandler {
	return NewDisabledDataTrieHandler()
}

// SaveDirtyData does nothing for this implementation
func (dtdt *disabledTrackableDataTrie) SaveDirtyData(_ common.Trie) ([]core.TrieData, error) {
	return make([]core.TrieData, 0), nil
}

// MigrateDataTrieLeaves does nothing for this implementation
func (dtdt *disabledTrackableDataTrie) MigrateDataTrieLeaves(_ vmcommon.ArgsMigrateDataTrieLeaves) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dtdt *disabledTrackableDataTrie) IsInterfaceNil() bool {
	return dtdt == nil
}
