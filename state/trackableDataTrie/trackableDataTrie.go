package trackableDataTrie

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("state/trackableDataTrie")

type dirtyData struct {
	value      []byte
	newVersion core.TrieNodeVersion
}

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type trackableDataTrie struct {
	dirtyData           map[string]dirtyData
	tr                  common.Trie
	hasher              hashing.Hasher
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
	identifier          []byte
}

// NewTrackableDataTrie returns an instance of trackableDataTrie
func NewTrackableDataTrie(
	identifier []byte,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	enableEpochsHandler common.EnableEpochsHandler,
) (*trackableDataTrie, error) {
	if check.IfNil(hasher) {
		return nil, state.ErrNilHasher
	}
	if check.IfNil(marshaller) {
		return nil, state.ErrNilMarshalizer
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, state.ErrNilEnableEpochsHandler
	}

	return &trackableDataTrie{
		tr:                  nil,
		hasher:              hasher,
		marshaller:          marshaller,
		dirtyData:           make(map[string]dirtyData),
		identifier:          identifier,
		enableEpochsHandler: enableEpochsHandler,
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

	// ok, not in cache, retrieve from trie
	if check.IfNil(tdaw.tr) {
		return nil, 0, state.ErrNilTrie
	}
	trieValue, depth, err := tdaw.retrieveValueFromTrie(key)
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
		value:      value,
		newVersion: core.GetVersionForNewData(tdaw.enableEpochsHandler),
	}

	tdaw.dirtyData[string(key)] = dataEntry
	return nil
}

// MigrateDataTrieLeaves migrates the data trie leaves from oldVersion to newVersion
func (tdaw *trackableDataTrie) MigrateDataTrieLeaves(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	if check.IfNil(tdaw.tr) {
		return state.ErrNilTrie
	}
	if check.IfNil(args.TrieMigrator) {
		return errorsCommon.ErrNilTrieMigrator
	}

	dtr, ok := tdaw.tr.(state.DataTrie)
	if !ok {
		return fmt.Errorf("invalid trie, type is %T", tdaw.tr)
	}

	err := dtr.CollectLeavesForMigration(args)
	if err != nil {
		return err
	}

	dataToBeMigrated := args.TrieMigrator.GetLeavesToBeMigrated()
	for _, leafData := range dataToBeMigrated {
		dataEntry := dirtyData{
			value:      leafData.Value,
			newVersion: args.NewVersion,
		}

		originalKey, err := tdaw.getOriginalKeyFromTrieData(leafData)
		if err != nil {
			return err
		}

		tdaw.dirtyData[string(originalKey)] = dataEntry
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
}

// DataTrie sets the internal data trie
func (tdaw *trackableDataTrie) DataTrie() common.DataTrieHandler {
	return tdaw.tr
}

// SaveDirtyData saved the dirty data to the trie
func (tdaw *trackableDataTrie) SaveDirtyData(mainTrie common.Trie) ([]core.TrieData, error) {
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

	dtr, ok := tdaw.tr.(state.DataTrie)
	if !ok {
		return nil, fmt.Errorf("invalid trie, type is %T", tdaw.tr)
	}

	return tdaw.updateTrie(dtr)
}

func (tdaw *trackableDataTrie) updateTrie(dtr state.DataTrie) ([]core.TrieData, error) {
	oldValues := make([]core.TrieData, len(tdaw.dirtyData))

	index := 0
	for key, dataEntry := range tdaw.dirtyData {
		oldVal, _, err := tdaw.retrieveValueFromTrie([]byte(key))
		if err != nil {
			return nil, err
		}
		oldValues[index] = oldVal

		err = tdaw.deleteOldEntryIfMigrated([]byte(key), dataEntry, oldVal)
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

func (tdaw *trackableDataTrie) retrieveValueFromTrie(key []byte) (core.TrieData, uint32, error) {
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

		return trieValue, depth, nil
	}

	newDataVersion := core.GetVersionForNewData(tdaw.enableEpochsHandler)
	keyForTrie := tdaw.getKeyForVersion(key, newDataVersion)

	trieValue := core.TrieData{
		Key:     keyForTrie,
		Value:   nil,
		Version: newDataVersion,
	}

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

func (tdaw *trackableDataTrie) deleteOldEntryIfMigrated(key []byte, newData dirtyData, oldEntry core.TrieData) error {
	if !tdaw.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		return nil
	}

	isMigration := oldEntry.Version == core.NotSpecified && newData.newVersion == core.AutoBalanceEnabled
	if isMigration && len(newData.value) != 0 {
		return tdaw.tr.Delete(key)
	}

	return nil
}

func (tdaw *trackableDataTrie) modifyTrie(key []byte, dataEntry dirtyData, oldVal core.TrieData, dtr state.DataTrie) error {
	if len(dataEntry.value) == 0 {
		return tdaw.deleteFromTrie(oldVal, key, dtr)
	}

	version := dataEntry.newVersion
	newKey := tdaw.getKeyForVersion(key, version)
	value, err := tdaw.getValueForVersion(key, dataEntry.value, version)
	if err != nil {
		return err
	}

	return dtr.UpdateWithVersion(newKey, value, version)
}

func (tdaw *trackableDataTrie) deleteFromTrie(oldVal core.TrieData, key []byte, dtr state.DataTrie) error {
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
