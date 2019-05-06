package state

import "github.com/ElrondNetwork/elrond-go-sandbox/data/trie"

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type TrackableDataTrie struct {
	originalData map[string][]byte
	dirtyData    map[string][]byte
	tr           trie.PatriciaMerkelTree
}

// NewTrackableDataTrie returns an instance of DataTrieTracker
func NewTrackableDataTrie(tr trie.PatriciaMerkelTree) *TrackableDataTrie {
	return &TrackableDataTrie{
		tr:           tr,
		originalData: make(map[string][]byte),
		dirtyData:    make(map[string][]byte),
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

	//search in dirty data cache
	data, found := tdaw.dirtyData[strKey]
	if found {
		return data, nil
	}

	//search in original data cache
	data, found = tdaw.originalData[strKey]
	if found {
		return data, nil
	}

	//ok, not in cache, retrieve from trie
	if tdaw.tr == nil {
		return nil, ErrNilTrie
	}
	data, err := tdaw.tr.Get(key)
	if err != nil {
		return nil, err
	}

	//got the value, put it originalData cache as the next fetch will run faster
	tdaw.originalData[string(key)] = data
	return data, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (tdaw *TrackableDataTrie) SaveKeyValue(key []byte, value []byte) {
	tdaw.dirtyData[string(key)] = value
}

// SetDataTrie sets the internal data trie
func (tdaw *TrackableDataTrie) SetDataTrie(tr trie.PatriciaMerkelTree) {
	tdaw.tr = tr
}

// DataTrie sets the internal data trie
func (tdaw *TrackableDataTrie) DataTrie() trie.PatriciaMerkelTree {
	return tdaw.tr
}
