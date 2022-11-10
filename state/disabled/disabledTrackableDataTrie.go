package disabled

import "github.com/ElrondNetwork/elrond-go/common"

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
func (dtdt *disabledTrackableDataTrie) SaveDirtyData(_ common.Trie) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dtdt *disabledTrackableDataTrie) IsInterfaceNil() bool {
	return dtdt == nil
}
