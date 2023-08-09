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
	migratedTrie        state.DataTrie
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
		migratedTrie:        nil,
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
	if check.IfNil(tdt.migratedTrie) {
		return nil, 0, fmt.Errorf("nil migrated trie")
	}

	hashedKey := tdt.hasher.Compute(string(key))
	valWithMetadata, depth, err := tdt.migratedTrie.Get(hashedKey)
	if err != nil {
		return nil, 0, err
	}

	if len(valWithMetadata) == 0 {
		return nil, 0, nil
	}

	val, err := tdt.getValueAutoBalanceVersion(valWithMetadata)
	if err != nil {
		return nil, 0, err
	}

	log.Trace("retrieve value from trie", "key", key, "value", val, "account", tdt.identifier)

	return val, depth, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (tdt *trackableDataTrie) SaveKeyValue(key []byte, value []byte) error {
	if uint64(len(value)) > core.MaxLeafSize {
		return data.ErrLeafSizeTooBig
	}

	dataEntry := dirtyData{
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
	for _, leafData := range dataToBeMigrated {
		dataEntry := dirtyData{
			value:      leafData.Value,
			newVersion: args.NewVersion,
		}

		originalKey, err := tdt.getOriginalKeyFromTrieData(leafData)
		if err != nil {
			return err
		}

		tdt.dirtyData[string(originalKey)] = dataEntry
	}

	return nil
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

	rootHash, err := tr.RootHash()
	if err != nil {
		log.Error("failed to get root hash", "err", err)
		return
	}

	migratedDataTrieRootHash, err := tdt.tr.GetStorageManager().Get(append(rootHash, state.MigratedRootHashKeySuffix...))
	if err != nil {
		log.Error("failed to get migrated root hash", "err", err)
		return
	}

	migratedDataTrie, err := tdt.getMigratedDataTrie(migratedDataTrieRootHash)
	if err != nil {
		log.Error("failed to get migrated trie", "err", err)
		return
	}

	tdt.migratedTrie = migratedDataTrie
}

func (tdt *trackableDataTrie) getMigratedDataTrie(rootHash []byte) (state.DataTrie, error) {
	migratedTrie, err := tdt.tr.Recreate(rootHash)
	if err != nil {
		return nil, err
	}

	migratedDataTrie, ok := migratedTrie.(state.DataTrie)
	if !ok {
		return nil, fmt.Errorf("invalid trie, type is %T", migratedTrie)
	}

	return migratedDataTrie, nil
}

// DataTrie sets the internal data trie
func (tdt *trackableDataTrie) DataTrie() common.DataTrieHandler {
	return tdt.tr
}

// SaveDirtyData saved the dirty data to the trie
func (tdt *trackableDataTrie) SaveDirtyData(mainTrie common.Trie) ([]core.TrieData, error) {
	if len(tdt.dirtyData) == 0 {
		return make([]core.TrieData, 0), nil
	}

	if check.IfNil(tdt.tr) {
		newDataTrie, err := mainTrie.Recreate(make([]byte, 0))
		if err != nil {
			return nil, err
		}

		migratedTrie, err := tdt.getMigratedDataTrie(make([]byte, 0))
		if err != nil {
			return nil, err
		}

		tdt.tr = newDataTrie
		tdt.migratedTrie = migratedTrie
	}

	dtr, ok := tdt.tr.(state.DataTrie)
	if !ok {
		return nil, fmt.Errorf("invalid trie, type is %T", tdt.tr)
	}

	migratedDataTrieRootHash, err := tdt.updateMigratedDataTrieAndCommit()
	if err != nil {
		return nil, err
	}

	oldValues, err := tdt.updateTrie(dtr)
	if err != nil {
		return nil, err
	}

	rootHash, err := tdt.tr.RootHash()
	if err != nil {
		return nil, err
	}

	err = mainTrie.GetStorageManager().Put(append(rootHash, state.MigratedRootHashKeySuffix...), migratedDataTrieRootHash)
	if err != nil {
		return nil, err
	}

	return oldValues, nil
}

func (tdt *trackableDataTrie) updateMigratedDataTrieAndCommit() ([]byte, error) {
	for key, dataEntry := range tdt.dirtyData {
		trieKey := tdt.getKeyForVersion([]byte(key), core.AutoBalanceEnabled)
		trieVal, err := tdt.getValueForVersion([]byte(key), dataEntry.value, core.AutoBalanceEnabled)
		if err != nil {
			return nil, err
		}

		err = tdt.migratedTrie.UpdateWithVersion(trieKey, trieVal, core.AutoBalanceEnabled)
		if err != nil {
			return nil, err
		}
	}

	err := tdt.migratedTrie.Commit()
	if err != nil {
		return nil, err
	}

	return tdt.migratedTrie.RootHash()
}

func (tdt *trackableDataTrie) updateTrie(dtr state.DataTrie) ([]core.TrieData, error) {
	oldValues := make([]core.TrieData, len(tdt.dirtyData))

	index := 0
	for key, dataEntry := range tdt.dirtyData {
		oldVal, _, err := tdt.retrieveValueFromTrie([]byte(key))
		if err != nil {
			return nil, err
		}
		oldValues[index] = oldVal

		err = tdt.deleteOldEntryIfMigrated([]byte(key), dataEntry, oldVal)
		if err != nil {
			return nil, err
		}

		err = tdt.modifyTrie([]byte(key), dataEntry, oldVal, dtr)
		if err != nil {
			return nil, err
		}

		index++
	}

	tdt.dirtyData = make(map[string]dirtyData)

	return oldValues, nil
}

func (tdt *trackableDataTrie) retrieveValueFromTrie(key []byte) (core.TrieData, uint32, error) {
	if tdt.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
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

func (tdt *trackableDataTrie) deleteOldEntryIfMigrated(key []byte, newData dirtyData, oldEntry core.TrieData) error {
	if !tdt.enableEpochsHandler.IsAutoBalanceDataTriesEnabled() {
		return nil
	}

	isMigration := oldEntry.Version == core.NotSpecified && newData.newVersion == core.AutoBalanceEnabled
	if isMigration && len(newData.value) != 0 {
		return tdt.tr.Delete(key)
	}

	return nil
}

func (tdt *trackableDataTrie) modifyTrie(key []byte, dataEntry dirtyData, oldVal core.TrieData, dtr state.DataTrie) error {
	if len(dataEntry.value) == 0 {
		return tdt.deleteFromTrie(oldVal, key, dtr)
	}

	version := dataEntry.newVersion
	newKey := tdt.getKeyForVersion(key, version)
	value, err := tdt.getValueForVersion(key, dataEntry.value, version)
	if err != nil {
		return err
	}

	return dtr.UpdateWithVersion(newKey, value, version)
}

func (tdt *trackableDataTrie) deleteFromTrie(oldVal core.TrieData, key []byte, dtr state.DataTrie) error {
	if len(oldVal.Value) == 0 {
		return nil
	}

	if oldVal.Version == core.AutoBalanceEnabled {
		return dtr.Delete(tdt.hasher.Compute(string(key)))
	}

	if oldVal.Version == core.NotSpecified {
		return dtr.Delete(key)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tdt *trackableDataTrie) IsInterfaceNil() bool {
	return tdt == nil
}
