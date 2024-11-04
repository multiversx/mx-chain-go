package trackableDataTrie

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
)

var log = logger.GetOrCreate("state/trackableDataTrie")

type dirtyData struct {
	index      int
	value      []byte
	newVersion core.TrieNodeVersion
}

// TrackableDataTrie wraps a PatriciaMerkelTrie adding modifying data capabilities
type trackableDataTrie struct {
	dirtyData             map[string]dirtyData
	tr                    common.Trie
	hasher                hashing.Hasher
	marshaller            marshal.Marshalizer
	enableEpochsHandler   common.EnableEpochsHandler
	identifier            []byte
	stateChangesCollector state.StateChangesCollector
}

// NewTrackableDataTrie returns an instance of trackableDataTrie
func NewTrackableDataTrie(
	identifier []byte,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	enableEpochsHandler common.EnableEpochsHandler,
	stateChangesCollector state.StateChangesCollector,
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
	if check.IfNil(stateChangesCollector) {
		return nil, state.ErrNilStateChangesCollector
	}

	err := core.CheckHandlerCompatibility(enableEpochsHandler, []core.EnableEpochFlag{
		common.AutoBalanceDataTriesFlag,
	})
	if err != nil {
		return nil, err
	}

	return &trackableDataTrie{
		tr:                    nil,
		hasher:                hasher,
		marshaller:            marshaller,
		dirtyData:             make(map[string]dirtyData),
		identifier:            identifier,
		enableEpochsHandler:   enableEpochsHandler,
		stateChangesCollector: stateChangesCollector,
	}, nil
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (tdt *trackableDataTrie) RetrieveValue(key []byte) ([]byte, uint32, error) {
	// search in dirty data cache
	if dataEntry, found := tdt.dirtyData[string(key)]; found {
		log.Trace("retrieve value from dirty data", "key", key, "value", dataEntry.value, "account", tdt.identifier)

		return dataEntry.value, 0, nil
	}

	// ok, not in cache, retrieve from trie
	if check.IfNil(tdt.tr) {
		return nil, 0, state.ErrNilTrie
	}
	trieValue, depth, err := tdt.retrieveValueFromTrie(key)
	if err != nil {
		return nil, depth, err
	}

	val, err := tdt.getValueWithoutMetadata(key, trieValue)
	if err != nil {
		return nil, depth, err
	}

	log.Trace("retrieve value from trie", "key", key, "value", val, "account", tdt.identifier)

	sc := &stateChange.StateChange{
		Type:        stateChange.Read,
		MainTrieKey: tdt.identifier,
		MainTrieVal: nil,
		DataTrieChanges: []*stateChange.DataTrieChange{
			{
				Type: stateChange.Read,
				Key:  key,
				Val:  val,
			},
		},
	}
	tdt.stateChangesCollector.AddStateChange(sc)

	return val, depth, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (tdt *trackableDataTrie) SaveKeyValue(key []byte, value []byte) error {
	if uint64(len(value)) > core.MaxLeafSize {
		return data.ErrLeafSizeTooBig
	}

	dataEntry := dirtyData{
		index:      tdt.getIndexForKey(key),
		value:      value,
		newVersion: core.GetVersionForNewData(tdt.enableEpochsHandler),
	}

	tdt.dirtyData[string(key)] = dataEntry
	return nil
}

// MigrateDataTrieLeaves migrates the data trie leaves from oldVersion to newVersion
func (tdt *trackableDataTrie) MigrateDataTrieLeaves(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	if check.IfNil(tdt.tr) {
		return state.ErrNilTrie
	}
	if check.IfNil(args.TrieMigrator) {
		return errorsCommon.ErrNilTrieMigrator
	}

	dtr, ok := tdt.tr.(state.DataTrie)
	if !ok {
		return fmt.Errorf("invalid trie, type is %T", tdt.tr)
	}

	err := dtr.CollectLeavesForMigration(args)
	if err != nil {
		return err
	}

	dataToBeMigrated := args.TrieMigrator.GetLeavesToBeMigrated()
	log.Debug("num leaves to be migrated", "num", len(dataToBeMigrated), "account", tdt.identifier)
	for _, leafData := range dataToBeMigrated {
		val, err := tdt.getValueWithoutMetadata(leafData.Key, leafData)
		if err != nil {
			return err
		}

		originalKey, err := tdt.getOriginalKeyFromTrieData(leafData)
		if err != nil {
			return err
		}

		dataEntry := dirtyData{
			index:      tdt.getIndexForKey(originalKey),
			value:      val,
			newVersion: args.NewVersion,
		}

		tdt.dirtyData[string(originalKey)] = dataEntry
	}

	return nil
}

func (tdt *trackableDataTrie) getIndexForKey(key []byte) int {
	existingVal, ok := tdt.dirtyData[string(key)]
	if ok {
		return existingVal.index
	}

	return len(tdt.dirtyData)
}

func (tdt *trackableDataTrie) getOriginalKeyFromTrieData(trieData core.TrieData) ([]byte, error) {
	if trieData.Version == core.AutoBalanceEnabled {
		valWithMetadata := &dataTrieValue.TrieLeafData{}
		err := tdt.marshaller.Unmarshal(valWithMetadata, trieData.Value)
		if err != nil {
			return nil, err
		}

		return valWithMetadata.Key, nil
	}

	return trieData.Key, nil
}

func (tdt *trackableDataTrie) getKeyForVersion(key []byte, version core.TrieNodeVersion) []byte {
	if version == core.AutoBalanceEnabled {
		return tdt.hasher.Compute(string(key))
	}

	return key
}

func (tdt *trackableDataTrie) getValueForVersion(key []byte, value []byte, version core.TrieNodeVersion) ([]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}

	if version == core.AutoBalanceEnabled {
		trieVal := &dataTrieValue.TrieLeafData{
			Value:   value,
			Key:     key,
			Address: tdt.identifier,
		}

		return tdt.marshaller.Marshal(trieVal)
	}

	identifier := append(key, tdt.identifier...)
	valueWithAppendedData := append(value, identifier...)

	return valueWithAppendedData, nil
}

// SetDataTrie sets the internal data trie
func (tdt *trackableDataTrie) SetDataTrie(tr common.Trie) {
	tdt.tr = tr
}

// DataTrie sets the internal data trie
func (tdt *trackableDataTrie) DataTrie() common.DataTrieHandler {
	return tdt.tr
}

// SaveDirtyData saved the dirty data to the trie
func (tdt *trackableDataTrie) SaveDirtyData(mainTrie common.Trie) ([]*stateChange.DataTrieChange, []core.TrieData, error) {
	if len(tdt.dirtyData) == 0 {
		return make([]*stateChange.DataTrieChange, 0), make([]core.TrieData, 0), nil
	}

	if check.IfNil(tdt.tr) {
		emptyRootHash := holders.NewDefaultRootHashesHolder(make([]byte, 0))
		newDataTrie, err := mainTrie.Recreate(emptyRootHash)
		if err != nil {
			return nil, nil, err
		}

		tdt.tr = newDataTrie
	}

	dtr, ok := tdt.tr.(state.DataTrie)
	if !ok {
		return nil, nil, fmt.Errorf("invalid trie, type is %T", tdt.tr)
	}

	return tdt.updateTrie(dtr)
}

func (tdt *trackableDataTrie) updateTrie(dtr state.DataTrie) ([]*stateChange.DataTrieChange, []core.TrieData, error) {
	oldValues := make([]core.TrieData, len(tdt.dirtyData))
	newData := make([]*stateChange.DataTrieChange, len(tdt.dirtyData))
	deletedKeys := make([]*stateChange.DataTrieChange, 0)

	index := 0
	for key, dataEntry := range tdt.dirtyData {
		oldVal, _, err := tdt.retrieveValueFromTrie([]byte(key))
		if err != nil {
			return nil, nil, err
		}
		oldValues[index] = oldVal

		wasDeleted, err := tdt.deleteOldEntryIfMigrated([]byte(key), dataEntry, oldVal)
		if err != nil {
			return nil, nil, err
		}

		if wasDeleted {
			deletedKeys = append(deletedKeys,
				&stateChange.DataTrieChange{
					Type: stateChange.Write,
					Key:  []byte(key),
					Val:  nil,
				},
			)
		}

		dataTrieKey, dataTrieVal, err := tdt.modifyTrie([]byte(key), dataEntry, oldVal, dtr)
		if err != nil {
			return nil, nil, err
		}

		index++

		isFirstMigration := oldVal.Version == core.NotSpecified && dataEntry.newVersion == core.AutoBalanceEnabled
		if isFirstMigration && len(dataTrieKey) != 0 {
			oldValues = append(oldValues, core.TrieData{
				Key:   dataTrieKey,
				Value: nil,
			})
		}

		if len(dataTrieKey) == 0 {
			continue
		}

		if dataEntry.index > len(newData)-1 {
			return nil, nil, fmt.Errorf("index out of range")
		}

		newData[dataEntry.index] = &stateChange.DataTrieChange{
			Type: stateChange.Write,
			Key:  dataTrieKey,
			Val:  dataTrieVal,
		}
	}

	tdt.dirtyData = make(map[string]dirtyData)

	stateChanges := make([]*stateChange.DataTrieChange, 0)
	for i := range newData {
		if newData[i] == nil {
			continue
		}
		if len(newData[i].Key) == 0 {
			continue
		}

		stateChanges = append(stateChanges, newData[i])
	}

	sort.Slice(deletedKeys, func(i, j int) bool {
		return bytes.Compare(deletedKeys[i].Key, deletedKeys[j].Key) < 0
	})
	stateChanges = append(stateChanges, deletedKeys...)

	return stateChanges, oldValues, nil
}

func (tdt *trackableDataTrie) retrieveValueFromTrie(key []byte) (core.TrieData, uint32, error) {
	if tdt.enableEpochsHandler.IsFlagEnabled(common.AutoBalanceDataTriesFlag) {
		hashedKey := tdt.hasher.Compute(string(key))
		valWithMetadata, depth, err := tdt.tr.Get(hashedKey)
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

	valWithMetadata, depth, err := tdt.tr.Get(key)
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

	newDataVersion := core.GetVersionForNewData(tdt.enableEpochsHandler)
	keyForTrie := tdt.getKeyForVersion(key, newDataVersion)

	trieValue := core.TrieData{
		Key:     keyForTrie,
		Value:   nil,
		Version: newDataVersion,
	}

	return trieValue, depth, nil
}

func (tdt *trackableDataTrie) getValueWithoutMetadata(key []byte, trieData core.TrieData) ([]byte, error) {
	if len(trieData.Value) == 0 {
		return nil, nil
	}

	if trieData.Version == core.AutoBalanceEnabled {
		return tdt.getValueAutoBalanceVersion(trieData.Value)
	}

	return tdt.getValueNotSpecifiedVersion(key, trieData.Value)
}

func (tdt *trackableDataTrie) getValueAutoBalanceVersion(val []byte) ([]byte, error) {
	dataTrieVal := &dataTrieValue.TrieLeafData{}
	err := tdt.marshaller.Unmarshal(dataTrieVal, val)
	if err != nil {
		return nil, err
	}

	return dataTrieVal.Value, nil
}

func (tdt *trackableDataTrie) getValueNotSpecifiedVersion(key []byte, val []byte) ([]byte, error) {
	tailLength := len(key) + len(tdt.identifier)
	trimmedValue, _ := common.TrimSuffixFromValue(val, tailLength)

	return trimmedValue, nil
}

func (tdt *trackableDataTrie) deleteOldEntryIfMigrated(key []byte, newData dirtyData, oldEntry core.TrieData) (bool, error) {
	if !tdt.enableEpochsHandler.IsFlagEnabled(common.AutoBalanceDataTriesFlag) {
		return false, nil
	}

	isMigration := oldEntry.Version == core.NotSpecified && newData.newVersion == core.AutoBalanceEnabled
	if isMigration && len(newData.value) != 0 {
		log.Trace("delete old entry if migrated", "key", key)
		return true, tdt.tr.Delete(key)
	}

	return false, nil
}

func (tdt *trackableDataTrie) modifyTrie(key []byte, dataEntry dirtyData, oldVal core.TrieData, dtr state.DataTrie) ([]byte, []byte, error) {
	version := dataEntry.newVersion
	newKey := tdt.getKeyForVersion(key, version)

	if len(dataEntry.value) == 0 {
		deletedKey, err := tdt.deleteFromTrie(oldVal, key, dtr)
		if err != nil {
			return nil, nil, err
		}

		return deletedKey, nil, nil
	}

	value, err := tdt.getValueForVersion(key, dataEntry.value, version)
	if err != nil {
		return nil, nil, err
	}

	err = dtr.UpdateWithVersion(newKey, value, version)
	if err != nil {
		return nil, nil, err
	}

	return newKey, value, nil
}

func (tdt *trackableDataTrie) deleteFromTrie(oldVal core.TrieData, key []byte, dtr state.DataTrie) ([]byte, error) {
	if len(oldVal.Value) == 0 {
		return nil, nil
	}

	if oldVal.Version == core.AutoBalanceEnabled {
		keyForTrie := tdt.hasher.Compute(string(key))
		return keyForTrie, dtr.Delete(keyForTrie)
	}

	if oldVal.Version == core.NotSpecified {
		return key, dtr.Delete(key)
	}

	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tdt *trackableDataTrie) IsInterfaceNil() bool {
	return tdt == nil
}
