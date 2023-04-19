package state

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	"github.com/multiversx/mx-chain-go/state/trieValuesCache"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type dirtyData struct {
	value   []byte
	version core.TrieNodeVersion
}

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type trackableDataTrie struct {
	dirtyData           map[string]dirtyData
	tr                  common.Trie
	hasher              hashing.Hasher
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
	identifier          []byte
	trieValuesCache     TrieValuesCacher
}

// NewTrackableDataTrie returns an instance of trackableDataTrie
func NewTrackableDataTrie(
	identifier []byte,
	tr common.Trie,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	enableEpochsHandler common.EnableEpochsHandler,
	trieValueCacher TrieValuesCacher,
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
	if check.IfNil(trieValueCacher) {
		return nil, ErrNilTrieValuesCacher
	}

	return &trackableDataTrie{
		tr:                  tr,
		hasher:              hasher,
		marshaller:          marshaller,
		dirtyData:           make(map[string]dirtyData),
		identifier:          identifier,
		enableEpochsHandler: enableEpochsHandler,
		trieValuesCache:     trieValueCacher,
	}, nil
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (tdaw *trackableDataTrie) RetrieveValue(key []byte) ([]byte, uint32, error) {
	// search in dirty data cache
	if dataEntry, found := tdaw.dirtyData[string(key)]; found {
		log.Trace("retrieve value from dirty data", "key", key, "value", dataEntry.value, "account", tdaw.identifier)
		return dataEntry.value, 0, nil
	}

	// search in trieValuesCache
	entry, ok := tdaw.trieValuesCache.Get(key)
	if ok {
		val, err := tdaw.getValueWithoutMetadata(key, entry)
		log.Trace("retrieve value from trie values cache", "key", key, "value", val, "account", tdaw.identifier)
		return val, 0, err
	}

	if check.IfNil(tdaw.tr) {
		return nil, 0, ErrNilTrie
	}

	// search in trie
	trieValue, depth, err := tdaw.retrieveValueFromTrieAndUpdateCache(key, tdaw.trieValuesCache)
	if err != nil {
		return nil, depth, err
	}

	val, err := tdaw.getValueWithoutMetadata(key, trieValue)
	if err != nil {
		return nil, depth, err
	}

	log.Trace("retrieve value from trie", "key", key, "value", val, "account", tdaw.identifier)

	return val, depth, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (tdaw *trackableDataTrie) SaveKeyValue(key []byte, value []byte) error {
	if uint64(len(value)) > core.MaxLeafSize {
		return data.ErrLeafSizeTooBig
	}

	dataEntry := dirtyData{
		value:   value,
		version: core.GetVersionForNewData(tdaw.enableEpochsHandler),
	}

	tdaw.dirtyData[string(key)] = dataEntry
	return nil
}

// MigrateDataTrieLeaves migrates the data trie leaves from oldVersion to newVersion
func (tdaw *trackableDataTrie) MigrateDataTrieLeaves(oldVersion core.TrieNodeVersion, newVersion core.TrieNodeVersion, trieMigrator vmcommon.DataTrieMigrator) error {
	if check.IfNil(tdaw.tr) {
		return ErrNilTrie
	}
	if check.IfNil(trieMigrator) {
		return errorsCommon.ErrNilTrieMigrator
	}

	dtr, ok := tdaw.tr.(dataTrie)
	if !ok {
		return fmt.Errorf("invalid trie, type is %T", tdaw.tr)
	}

	err := dtr.CollectLeavesForMigration(oldVersion, newVersion, trieMigrator)
	if err != nil {
		return err
	}

	dataToBeMigrated := trieMigrator.GetLeavesToBeMigrated()
	for _, leafData := range dataToBeMigrated {
		newDataEntry := dirtyData{
			value:   leafData.Value,
			version: newVersion,
		}

		originalKey, err := tdaw.getOriginalKeyFromTrieData(leafData)
		if err != nil {
			return err
		}

		tdaw.dirtyData[string(originalKey)] = newDataEntry

		oldTrieVal := core.TrieData{
			Key:     leafData.Key,
			Value:   leafData.Value,
			Version: leafData.Version,
		}

		tdaw.trieValuesCache.Put(originalKey, oldTrieVal)
	}

	return nil
}

func (tdaw *trackableDataTrie) getOriginalKeyFromTrieData(trieData core.TrieData) ([]byte, error) {
	if trieData.Version == core.AutoBalanceEnabled {
		valWithMetadata := &dataTrieValue.TrieLeafData{}
		err := tdaw.marshaller.Unmarshal(valWithMetadata, trieData.Value)
		if err != nil {
			return nil, err
		}

		return valWithMetadata.Key, nil
	}

	return trieData.Key, nil
}

func (tdaw *trackableDataTrie) getKeyForVersion(key []byte, version core.TrieNodeVersion) []byte {
	if version == core.AutoBalanceEnabled {
		return tdaw.hasher.Compute(string(key))
	}

	return key
}

func (tdaw *trackableDataTrie) getValueForVersion(key []byte, value []byte, version core.TrieNodeVersion) ([]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}

	if version == core.AutoBalanceEnabled {
		trieVal := &dataTrieValue.TrieLeafData{
			Value:   value,
			Key:     key,
			Address: tdaw.identifier,
		}

		return tdaw.marshaller.Marshal(trieVal)
	}

	identifier := append(key, tdaw.identifier...)
	valueWithAppendedData := append(value, identifier...)

	return valueWithAppendedData, nil
}

// SetDataTrie sets the internal data trie
func (tdaw *trackableDataTrie) SetDataTrie(tr common.Trie) {
	tdaw.tr = tr

	tdaw.trieValuesCache.Clean()
}

// DataTrie sets the internal data trie
func (tdaw *trackableDataTrie) DataTrie() common.DataTrieHandler {
	return tdaw.tr
}

// SaveDirtyData saved the dirty data to the trie
func (tdaw *trackableDataTrie) SaveDirtyData(mainTrie common.Trie) ([]core.TrieData, error) {
	// TODO investigate if it is better to only clean the cache on data trie commit.
	// If so, treat the case when the data is modified for a cached value, and the case of a reverted tx
	defer tdaw.trieValuesCache.Clean()

	if len(tdaw.dirtyData) == 0 {
		return make([]core.TrieData, 0), nil
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

	return tdaw.updateTrie(dtr)
}

func (tdaw *trackableDataTrie) updateTrie(dtr dataTrie) ([]core.TrieData, error) {
	oldValues := make([]core.TrieData, len(tdaw.dirtyData))

	index := 0
	for key, dataEntry := range tdaw.dirtyData {
		oldVal, err := tdaw.getOldValue([]byte(key))
		if err != nil {
			return nil, err
		}
		oldValues[index] = oldVal

		err = tdaw.deleteOldEntryIfMigrated([]byte(key), oldVal.Version, dataEntry)
		if err != nil {
			return nil, err
		}

		err = tdaw.modifyTrie([]byte(key), dataEntry, oldVal, dtr)
		if err != nil {
			return nil, err
		}

		index++
	}

	tdaw.dirtyData = make(map[string]dirtyData)

	return oldValues, nil
}

func (tdaw *trackableDataTrie) getOldValue(key []byte) (core.TrieData, error) {
	entry, ok := tdaw.trieValuesCache.Get(key)
	if ok {
		log.Trace("trieValuesCache hit", "key", key, "account", tdaw.identifier)
		return entry, nil
	}

	log.Trace("trieValuesCache miss", "key", key, "account", tdaw.identifier)

	val, _, err := tdaw.retrieveValueFromTrieAndUpdateCache(key, trieValuesCache.NewDisabledTrieValuesCache())
	return val, err
}

func (tdaw *trackableDataTrie) retrieveValueFromTrieAndUpdateCache(key []byte, trieValCache TrieValuesCacher) (core.TrieData, uint32, error) {
	if tdaw.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		hashedKey := tdaw.hasher.Compute(string(key))
		valWithMetadata, depth, err := tdaw.tr.Get(hashedKey)
		if err != nil {
			return core.TrieData{}, 0, err
		}
		if len(valWithMetadata) != 0 {
			trieValue := core.TrieData{
				Key:     hashedKey,
				Value:   valWithMetadata,
				Version: core.AutoBalanceEnabled,
			}

			trieValCache.Put(key, trieValue)

			return trieValue, depth, nil
		}
	}

	valWithMetadata, depth, err := tdaw.tr.Get(key)
	if err != nil {
		return core.TrieData{}, 0, err
	}
	if len(valWithMetadata) != 0 {
		trieValue := core.TrieData{
			Key:     key,
			Value:   valWithMetadata,
			Version: core.NotSpecified,
		}

		trieValCache.Put(key, trieValue)

		return trieValue, depth, nil
	}

	newDataVersion := core.GetVersionForNewData(tdaw.enableEpochsHandler)
	keyForTrie := tdaw.getKeyForVersion(key, newDataVersion)

	trieValue := core.TrieData{
		Key:     keyForTrie,
		Value:   nil,
		Version: newDataVersion,
	}

	trieValCache.Put(key, trieValue)

	return trieValue, depth, nil
}

func (tdaw *trackableDataTrie) getValueWithoutMetadata(key []byte, trieData core.TrieData) ([]byte, error) {
	if len(trieData.Value) == 0 {
		return nil, nil
	}

	if trieData.Version == core.AutoBalanceEnabled {
		return tdaw.getValueAutoBalanceVersion(trieData.Value)
	}

	return tdaw.getValueNotSpecifiedVersion(key, trieData.Value)
}

func (tdaw *trackableDataTrie) getValueAutoBalanceVersion(val []byte) ([]byte, error) {
	dataTrieVal := &dataTrieValue.TrieLeafData{}
	err := tdaw.marshaller.Unmarshal(dataTrieVal, val)
	if err != nil {
		return nil, err
	}

	return dataTrieVal.Value, nil
}

func (tdaw *trackableDataTrie) getValueNotSpecifiedVersion(key []byte, val []byte) ([]byte, error) {
	tailLength := len(key) + len(tdaw.identifier)
	trimmedValue, _ := common.TrimSuffixFromValue(val, tailLength)

	return trimmedValue, nil
}

func (tdaw *trackableDataTrie) deleteOldEntryIfMigrated(key []byte, oldVersion core.TrieNodeVersion, newData dirtyData) error {
	if !tdaw.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		return nil
	}

	isMigration := oldVersion == core.NotSpecified && newData.version == core.AutoBalanceEnabled
	if isMigration && len(newData.value) != 0 {
		return tdaw.tr.Delete(key)
	}

	return nil
}

func (tdaw *trackableDataTrie) modifyTrie(key []byte, dataEntry dirtyData, oldVal core.TrieData, dtr dataTrie) error {
	if len(dataEntry.value) == 0 {
		return tdaw.deleteFromTrie(oldVal, key, dtr)
	}

	version := dataEntry.version
	newKey := tdaw.getKeyForVersion(key, version)
	value, err := tdaw.getValueForVersion(key, dataEntry.value, version)
	if err != nil {
		return err
	}

	return dtr.UpdateWithVersion(newKey, value, version)
}

func (tdaw *trackableDataTrie) deleteFromTrie(oldVal core.TrieData, key []byte, dtr dataTrie) error {
	if len(oldVal.Value) == 0 {
		return nil
	}

	if oldVal.Version == core.AutoBalanceEnabled {
		return dtr.Delete(tdaw.hasher.Compute(string(key)))
	}

	if oldVal.Version == core.NotSpecified {
		return dtr.Delete(key)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tdaw *trackableDataTrie) IsInterfaceNil() bool {
	return tdaw == nil
}
