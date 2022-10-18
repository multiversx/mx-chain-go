package state

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
)

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type trackableDataTrie struct {
	dirtyData  map[string][]byte
	tr         common.Trie
	identifier []byte
}

// NewTrackableDataTrie returns an instance of trackableDataTrie
func NewTrackableDataTrie(identifier []byte, tr common.Trie) *trackableDataTrie {
	return &trackableDataTrie{
		tr:         tr,
		dirtyData:  make(map[string][]byte),
		identifier: identifier,
	}
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (tdaw *trackableDataTrie) RetrieveValue(key []byte) ([]byte, error) {
	tailLength := len(key) + len(tdaw.identifier)

	// search in dirty data cache
	if value, found := tdaw.dirtyData[string(key)]; found {
		log.Trace("retrieve value from dirty data", "key", key, "value", value)
		return trimValue(value, tailLength)
	}

	// ok, not in cache, retrieve from trie
	if tdaw.tr == nil {
		return nil, ErrNilTrie
	}
	value, _, err := tdaw.tr.Get(key)
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
func (tdaw *trackableDataTrie) SaveKeyValue(key []byte, value []byte) error {
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
func (tdaw *trackableDataTrie) SetDataTrie(tr common.Trie) {
	tdaw.tr = tr
}

// DataTrie sets the internal data trie
func (tdaw *trackableDataTrie) DataTrie() common.DataTrieHandler {
	return tdaw.tr
}

// SaveDirtyData saved the dirty data to the trie
func (tdaw *trackableDataTrie) SaveDirtyData(mainTrie common.Trie) (map[string][]byte, error) {
	if len(tdaw.dirtyData) == 0 {
		return map[string][]byte{}, nil
	}

	if check.IfNil(tdaw.tr) {
		newDataTrie, err := mainTrie.Recreate(make([]byte, 0))
		if err != nil {
			return nil, err
		}

		tdaw.tr = newDataTrie
	}

	oldValues := make(map[string][]byte)

	for k, v := range tdaw.dirtyData {
		val, _, err := tdaw.tr.Get([]byte(k))
		if err != nil {
			return oldValues, err
		}

		oldValues[k] = val

		err = tdaw.tr.Update([]byte(k), v)
		if err != nil {
			return oldValues, err
		}
	}

	tdaw.dirtyData = make(map[string][]byte)
	return oldValues, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tdaw *trackableDataTrie) IsInterfaceNil() bool {
	return tdaw == nil
}
