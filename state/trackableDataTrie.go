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
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type optionalVersion struct {
	version      core.TrieNodeVersion
	isValueKnown bool
}

type dirtyData struct {
	value      []byte
	oldVersion optionalVersion
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
		log.Trace("retrieve value from dirty data", "key", key, "value", dataEntry.value)
		return dataEntry.value, 0, nil
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

	dataEntry := dirtyData{
		value: value,
		oldVersion: optionalVersion{
			isValueKnown: false,
		},
		newVersion: core.GetVersionForNewData(tdaw.enableEpochsHandler),
	}

	tdaw.dirtyData[string(key)] = dataEntry
	return nil
}

// MigrateDataTrieLeaves migrates the data trie leaves from oldVersion to newVersion
func (tdaw *trackableDataTrie) MigrateDataTrieLeaves(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	if check.IfNil(tdaw.tr) {
		return ErrNilTrie
	}
	if check.IfNil(args.TrieMigrator) {
		return errorsCommon.ErrNilTrieMigrator
	}

	dtr, ok := tdaw.tr.(dataTrie)
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
			value: leafData.Value,
			oldVersion: optionalVersion{
				version:      leafData.Version,
				isValueKnown: true,
			},
			newVersion: args.NewVersion,
		}

		tdaw.dirtyData[string(leafData.Key)] = dataEntry
	}

	return nil
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
		// TODO cache old value if it was previously retrieved from the trie
		oldVal := tdaw.getOldValue([]byte(key), dataEntry)
		oldValues[index] = oldVal

		err := tdaw.deleteOldEntryIfMigrated([]byte(key), dataEntry, oldVal)
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

func (tdaw *trackableDataTrie) getOldValue(key []byte, dataEntry dirtyData) core.TrieData {
	if dataEntry.oldVersion.isValueKnown {
		return core.TrieData{
			Key:     key,
			Value:   dataEntry.value,
			Version: dataEntry.oldVersion.version,
		}
	}

	if tdaw.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		hashedKey := tdaw.hasher.Compute(string(key))
		oldVal, _, err := tdaw.tr.Get(hashedKey)
		if err == nil && len(oldVal) != 0 {
			return core.TrieData{
				Key:     hashedKey,
				Value:   oldVal,
				Version: core.AutoBalanceEnabled,
			}
		}
	}

	oldVal, _, err := tdaw.tr.Get(key)
	if err == nil && len(oldVal) != 0 {
		return core.TrieData{
			Key:     key,
			Value:   oldVal,
			Version: core.NotSpecified,
		}
	}

	newDataVersion := core.GetVersionForNewData(tdaw.enableEpochsHandler)
	return core.TrieData{
		Key:     tdaw.getKeyForVersion(key, newDataVersion),
		Value:   nil,
		Version: newDataVersion,
	}
}

func (tdaw *trackableDataTrie) deleteOldEntryIfMigrated(key []byte, newData dirtyData, oldEntry core.TrieData) error {
	if !tdaw.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		return nil
	}

	if oldEntry.Version == core.NotSpecified && newData.newVersion == core.AutoBalanceEnabled {
		return tdaw.tr.Delete(key)
	}

	return nil
}

func (tdaw *trackableDataTrie) modifyTrie(key []byte, dataEntry dirtyData, oldVal core.TrieData, dtr dataTrie) error {
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
