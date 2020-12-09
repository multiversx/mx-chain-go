package state

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type TrackableDataTrie struct {
	dirtyData  map[string][]byte
	tr         data.Trie
	identifier []byte
}

// NewTrackableDataTrie returns an instance of DataTrieTracker
func NewTrackableDataTrie(identifier []byte, tr data.Trie) *TrackableDataTrie {
	return &TrackableDataTrie{
		tr:         tr,
		dirtyData:  make(map[string][]byte),
		identifier: identifier,
	}
}

// ClearDataCaches empties the dirtyData map and original map
func (tdaw *TrackableDataTrie) ClearDataCaches() {
	tdaw.dirtyData = make(map[string][]byte)
}

// DirtyData returns the map of (key, value) pairs that contain the data needed to be saved in the data trie
func (tdaw *TrackableDataTrie) DirtyData() map[string][]byte {
	return tdaw.dirtyData
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// data must have been retrieved from its trie
func (tdaw *TrackableDataTrie) RetrieveValue(key []byte) ([]byte, error) {
	tailLength := len(key) + len(tdaw.identifier)

	//search in dirty data cache
	if value, found := tdaw.dirtyData[string(key)]; found {
		log.Trace("retrieve value from dirty data", "key", key, "value", value)
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
	log.Trace("retrieve value from trie", "key", key, "value", value)
	value, _ = trimValue(value, tailLength)

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
func (tdaw *TrackableDataTrie) SaveKeyValue(key []byte, value []byte) error {
	var identifier []byte
	lenValue := uint64(len(value))
	if lenValue > core.MaxLeafSize {
		return data.ErrLeafSizeTooBig
	}

	if lenValue != 0 {
		identifier = append(key, tdaw.identifier...)
	}

	tdaw.dirtyData[string(key)] = append(value, identifier...)
	return nil
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
	return tdaw == nil
}
