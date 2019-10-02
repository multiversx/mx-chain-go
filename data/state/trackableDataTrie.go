package state

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type TrackableDataTrie struct {
	originalData map[string][]byte
	dirtyData    map[string][]byte
	tr           data.Trie
	identifier   []byte
}

// NewTrackableDataTrie returns an instance of DataTrieTracker
func NewTrackableDataTrie(identifier []byte, tr data.Trie) *TrackableDataTrie {
	return &TrackableDataTrie{
		tr:           tr,
		originalData: make(map[string][]byte),
		dirtyData:    make(map[string][]byte),
		identifier:   identifier,
	}
}

// ClearDataCaches empties the dirtyData map and original map
func (tdaw *TrackableDataTrie) ClearDataCaches() {
	tdaw.dirtyData = make(map[string][]byte)
	tdaw.originalData = make(map[string][]byte)
}

// DirtyData returns the map of (key, value) pairs that contain the data needed to be saved in the data trie
func (tdaw *TrackableDataTrie) DirtyData() map[string][]byte {
	return tdaw.dirtyData
}

// OriginalValue returns the value for a key stored in originalData map which is acting like a cache
func (tdaw *TrackableDataTrie) OriginalValue(key []byte) []byte {
	return tdaw.originalData[string(key)]
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (tdaw *TrackableDataTrie) RetrieveValue(key []byte) ([]byte, error) {
	strKey := string(key)
	tailLength := len(key) + len(tdaw.identifier)

	//search in dirty data cache
	value, found := tdaw.dirtyData[strKey]
	if found {
		return trimValue(value, tailLength)
	}

	//search in original data cache
	value, found = tdaw.originalData[strKey]
	if found {
		return trimValue(value, tailLength)
	}

	//ok, not in cache, retrieve from trie
	if tdaw.tr == nil {
		return nil, ErrNilTrie
	}
	value, err := tdaw.tr.Get(key)
	if err != nil {
		return nil, err
	}

	value, _ = trimValue(value, tailLength)

	//got the value, put it originalData cache as the next fetch will run faster
	tdaw.originalData[string(key)] = value
	return value, nil
}

func trimValue(value []byte, tailLength int) ([]byte, error) {
	dataLength := len(value) - tailLength
	if dataLength < 0 {
		return nil, ErrNegativeValue
	}

	return value[:dataLength], nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (tdaw *TrackableDataTrie) SaveKeyValue(key []byte, value []byte) {
	identifier := append(key, tdaw.identifier...)
	tdaw.dirtyData[string(key)] = append(value, identifier...)
}

// SetDataTrie sets the internal data trie
func (tdaw *TrackableDataTrie) SetDataTrie(tr data.Trie) {
	tdaw.tr = tr
}

// DataTrie sets the internal data trie
func (tdaw *TrackableDataTrie) DataTrie() data.Trie {
	return tdaw.tr
}

// IsInterfaceNil returns true if there is no value under the interface
func (tdaw *TrackableDataTrie) IsInterfaceNil() bool {
	if tdaw == nil {
		return true
	}
	return false
}
