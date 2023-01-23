package state

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
)

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type trackableDataTrie struct {
	dirtyData           map[string][]byte
	tr                  common.Trie
	hasher              hashing.Hasher
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
	identifier          []byte
}

// NewTrackableDataTrie returns an instance of trackableDataTrie
func NewTrackableDataTrie(
	identifier []byte,
	tr common.Trie,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	enableEpochsHandler common.EnableEpochsHandler,
) (*trackableDataTrie, error) {
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, ErrNilEnableEpochsHandler
	}

	return &trackableDataTrie{
		tr:                  tr,
		hasher:              hasher,
		marshaller:          marshaller,
		dirtyData:           make(map[string][]byte),
		identifier:          identifier,
		enableEpochsHandler: enableEpochsHandler,
	}, nil
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (tdaw *trackableDataTrie) RetrieveValue(key []byte) ([]byte, uint32, error) {
	// search in dirty data cache
	if value, found := tdaw.dirtyData[string(key)]; found {
		log.Trace("retrieve value from dirty data", "key", key, "value", value)
		return value, 0, nil
	}

	// ok, not in cache, retrieve from trie
	if check.IfNil(tdaw.tr) {
		return nil, 0, ErrNilTrie
	}
	return tdaw.retrieveVal(string(key))
}

func (tdaw *trackableDataTrie) retrieveVal(key string) ([]byte, uint32, error) {
	if !tdaw.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		return tdaw.retrieveValV1([]byte(key))
	}

	val, depth, err := tdaw.tr.Get(tdaw.hasher.Compute(key))
	if err != nil {
		return nil, 0, err
	}

	if len(val) == 0 {
		return tdaw.retrieveValV1([]byte(key))
	}

	dataTrieVal := &dataTrieValue.TrieLeafData{}
	err = tdaw.marshaller.Unmarshal(dataTrieVal, val)
	if err != nil {
		return nil, depth, err
	}

	log.Trace("retrieve value from trie V2", "key", key, "value", dataTrieVal.Value)
	return dataTrieVal.Value, depth, nil
}

func (tdaw *trackableDataTrie) retrieveValV1(key []byte) ([]byte, uint32, error) {
	val, depth, err := tdaw.tr.Get(key)
	if err != nil {
		return nil, 0, err
	}

	tailLength := len(key) + len(tdaw.identifier)
	value, _ := common.TrimSuffixFromValue(val, tailLength)
	log.Trace("retrieve value from trie V1", "key", key, "value", value)
	return value, depth, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (tdaw *trackableDataTrie) SaveKeyValue(key []byte, value []byte) error {
	if uint64(len(value)) > core.MaxLeafSize {
		return data.ErrLeafSizeTooBig
	}

	tdaw.dirtyData[string(key)] = value
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
func (tdaw *trackableDataTrie) SaveDirtyData(mainTrie common.Trie) ([]common.TrieData, error) {
	if len(tdaw.dirtyData) == 0 {
		return make([]common.TrieData, 0), nil
	}

	if check.IfNil(tdaw.tr) {
		newDataTrie, err := mainTrie.Recreate(make([]byte, 0))
		if err != nil {
			return nil, err
		}

		tdaw.tr = newDataTrie
	}

	dtr, ok := tdaw.tr.(dataTrie)
	if !ok {
		return nil, fmt.Errorf("invalid trie, type is %T", tdaw.tr)
	}

	if tdaw.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		return tdaw.updateTrieWithAutoBalancing(dtr)
	}

	return tdaw.updateTrieV1(dtr)
}

func (tdaw *trackableDataTrie) updateTrieV1(selfDataTrie dataTrie) ([]common.TrieData, error) {
	oldValues := make([]common.TrieData, len(tdaw.dirtyData))

	index := 0
	for key, val := range tdaw.dirtyData {
		oldVal, _, err := tdaw.tr.Get([]byte(key))
		if err != nil {
			return nil, err
		}

		oldEntry := common.TrieData{
			Key:     []byte(key),
			Value:   oldVal,
			Version: common.NotSpecified,
		}
		oldValues[index] = oldEntry

		var identifier []byte
		if len(val) != 0 {
			identifier = append([]byte(key), tdaw.identifier...)
		}

		valueWithAppendedData := append(val, identifier...)

		err = selfDataTrie.UpdateWithVersion([]byte(key), valueWithAppendedData, common.NotSpecified)
		if err != nil {
			return nil, err
		}

		index++
	}

	tdaw.dirtyData = make(map[string][]byte)
	return oldValues, nil
}

// TODO refactor to make the migration more generic. This code should be able to migrate between specified versions.

func (tdaw *trackableDataTrie) updateTrieWithAutoBalancing(dtr dataTrie) ([]common.TrieData, error) {
	oldValues := make([]common.TrieData, len(tdaw.dirtyData))

	index := 0
	for key, val := range tdaw.dirtyData {
		oldEntry, err := tdaw.getOldKeyAndValWithCleanup(key)
		if err != nil {
			return nil, err
		}

		oldValues[index] = oldEntry

		err = tdaw.updateValInTrieWithAutoBalancing([]byte(key), val, dtr)
		if err != nil {
			return nil, err
		}

		index++
	}

	tdaw.dirtyData = make(map[string][]byte)
	return oldValues, nil
}

func (tdaw *trackableDataTrie) getOldKeyAndValWithCleanup(key string) (common.TrieData, error) {
	hashedKey := tdaw.hasher.Compute(key)

	oldVal, _, err := tdaw.tr.Get(hashedKey)
	if err == nil && len(oldVal) != 0 {
		return common.TrieData{
			Key:     hashedKey,
			Value:   oldVal,
			Version: common.AutoBalanceEnabled,
		}, nil
	}

	oldVal, _, err = tdaw.tr.Get([]byte(key))
	if err != nil {
		return common.TrieData{}, err
	}

	if len(oldVal) == 0 {
		return common.TrieData{
			Key:     hashedKey,
			Value:   nil,
			Version: common.NotSpecified,
		}, nil
	}

	err = tdaw.tr.Delete([]byte(key))
	if err != nil {
		return common.TrieData{}, err
	}

	return common.TrieData{
		Key:     []byte(key),
		Value:   oldVal,
		Version: common.NotSpecified,
	}, nil
}

func (tdaw *trackableDataTrie) updateValInTrieWithAutoBalancing(key []byte, val []byte, selfDataTrie dataTrie) error {
	if len(val) == 0 {
		return tdaw.tr.Delete(tdaw.hasher.Compute(string(key)))
	}

	trieVal := &dataTrieValue.TrieLeafData{
		Value:   val,
		Key:     key,
		Address: tdaw.identifier,
	}

	serializedTrieVal, err := tdaw.marshaller.Marshal(trieVal)
	if err != nil {
		return err
	}

	return selfDataTrie.UpdateWithVersion(tdaw.hasher.Compute(string(key)), serializedTrieVal, common.AutoBalanceEnabled)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tdaw *trackableDataTrie) IsInterfaceNil() bool {
	return tdaw == nil
}
