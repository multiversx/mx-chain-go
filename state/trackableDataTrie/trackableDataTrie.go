package trackableDataTrie

import (
	"bytes"
	"fmt"
	"sort"
	"time"

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
	"github.com/multiversx/mx-chain-go/common/stateAnalysis"
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
	dirtyData              map[string]dirtyData
	tr                     common.Trie
	hasher                 hashing.Hasher
	marshaller             marshal.Marshalizer
	enableEpochsHandler    common.EnableEpochsHandler
	identifier             []byte
	stateAccessesCollector state.StateAccessesCollector
}

// NewTrackableDataTrie returns an instance of trackableDataTrie
func NewTrackableDataTrie(
	identifier []byte,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	enableEpochsHandler common.EnableEpochsHandler,
	stateAccessesCollector state.StateAccessesCollector,
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
	if check.IfNil(stateAccessesCollector) {
		return nil, state.ErrNilStateAccessesCollector
	}

	err := core.CheckHandlerCompatibility(enableEpochsHandler, []core.EnableEpochFlag{
		common.AutoBalanceDataTriesFlag,
		common.FixDataTrieLeafValidationFlag,
	})
	if err != nil {
		return nil, err
	}

	return &trackableDataTrie{
		tr:                     nil,
		hasher:                 hasher,
		marshaller:             marshaller,
		dirtyData:              make(map[string]dirtyData),
		identifier:             identifier,
		enableEpochsHandler:    enableEpochsHandler,
		stateAccessesCollector: stateAccessesCollector,
	}, nil
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (tdt *trackableDataTrie) RetrieveValue(key []byte) ([]byte, uint32, error) {
	stateAnalysis.Inc(stateAnalysis.CDtRetrieve)
	// search in dirty data cache
	if dataEntry, found := tdt.dirtyData[string(key)]; found {
		log.Trace("retrieve value from dirty data", "key", key, "value", dataEntry.value, "account", tdt.identifier)
		stateAnalysis.Inc(stateAnalysis.CDtRetrieveDirtyHit)

		return dataEntry.value, 0, nil
	}

	// ok, not in cache, retrieve from trie
	if check.IfNil(tdt.tr) {
		return nil, 0, state.ErrNilTrie
	}
	retrieveStart := stateAnalysis.Now()
	trieValue, depth, err := tdt.retrieveValueFromTrie(key, false)
	stateAnalysis.AddSince(stateAnalysis.CDtRetrieveNs, retrieveStart)
	if err != nil {
		return nil, depth, err
	}

	val, err := tdt.getValueWithoutMetadata(key, trieValue)
	if err != nil {
		return nil, depth, err
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
		index:      tdt.getIndexForKey(key),
		value:      value,
		newVersion: core.GetVersionForNewData(tdt.enableEpochsHandler),
	}

	stateAnalysis.Inc(stateAnalysis.CDtSaveKeyValue)
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
		tdt.collectReadOperation(leafData.Key, leafData.Value, uint32(leafData.Version))
		originalKey, err := tdt.getOriginalKeyFromTrieData(leafData)
		if err != nil {
			return err
		}

		val, err := tdt.getValueWithoutMetadata(originalKey, leafData)
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
		prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindData)
		newDataTrie, err := mainTrie.Recreate(emptyRootHash)
		stateAnalysis.SetTrieKind(prevKind)
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

func (tdt *trackableDataTrie) rollbackAppliedUpdates(dtr state.DataTrie, oldValues []core.TrieData) {
	stateAnalysis.Inc(stateAnalysis.CDtRollback)
	var trieUpdate core.TrieData
	var err error

	for i := len(oldValues) - 1; i >= 0; i-- {
		trieUpdate = oldValues[i]
		err = dtr.UpdateWithVersion(trieUpdate.Key, trieUpdate.Value, trieUpdate.Version)
		if err != nil {
			log.Error("could not apply rollback updates", "err", err, "key", trieUpdate.Key, "account", tdt.identifier)
		}
	}
}

func (tdt *trackableDataTrie) updateTrie(dtr state.DataTrie) ([]*stateChange.DataTrieChange, []core.TrieData, error) {
	if stateAnalysis.Enabled() {
		stateAnalysis.Inc(stateAnalysis.CDtUpdateTrieCalls)
		stateAnalysis.Add(stateAnalysis.CDtUpdateDirtyKeys, int64(len(tdt.dirtyData)))
		stateAnalysis.Max(stateAnalysis.MDtDirtyKeys, int64(len(tdt.dirtyData)))
		defer stateAnalysis.AddSince(stateAnalysis.CDtUpdateNs, time.Now())
	}
	oldValues := make([]core.TrieData, len(tdt.dirtyData))
	newData := make([]*stateChange.DataTrieChange, len(tdt.dirtyData))
	deletedKeys := make([]*stateChange.DataTrieChange, 0)
	trieUpdates := make([]core.TrieData, 0)

	keys := make([]string, 0, len(tdt.dirtyData))
	for key := range tdt.dirtyData {
		keys = append(keys, key)
	}
	if len(keys) > 1 && tdt.enableEpochsHandler.IsFlagEnabled(common.FixDataTrieLeafValidationFlag) {
		// Interacting migrations must not depend on Go map iteration order.
		sort.Strings(keys)
	}

	index := 0
	for _, key := range keys {
		dataEntry := tdt.dirtyData[key]
		writesHashedKey := len(dataEntry.value) != 0 && dataEntry.newVersion == core.AutoBalanceEnabled
		oldReadStart := stateAnalysis.Now()
		oldVal, _, err := tdt.retrieveValueFromTrie([]byte(key), writesHashedKey)
		stateAnalysis.AddSince(stateAnalysis.CDtUpdateOldReadNs, oldReadStart)
		if err != nil {
			tdt.rollbackAppliedUpdates(dtr, trieUpdates)
			return nil, nil, err
		}
		oldValues[index] = oldVal
		tdt.classifyUpdate([]byte(key), dataEntry, oldVal)
		if oldVal.Version == core.AutoBalanceEnabled && len(oldVal.Value) != 0 {
			_, decodeErr := tdt.getValueAutoBalanceVersion([]byte(key), oldVal.Value)
			if decodeErr != nil && tdt.enableEpochsHandler.IsFlagEnabled(common.FixDataTrieLeafValidationFlag) {
				tdt.rollbackAppliedUpdates(dtr, trieUpdates)
				return nil, nil, decodeErr
			}
		}

		wasDeleted, err := tdt.deleteOldEntryIfMigrated([]byte(key), dataEntry, oldVal)
		if err != nil {
			tdt.rollbackAppliedUpdates(dtr, trieUpdates)
			return nil, nil, err
		}

		if wasDeleted {
			trieUpdates = append(trieUpdates, oldVal)
			originalVal := tdt.getValueNotSpecifiedVersion([]byte(key), oldVal.Value)

			deletedKeys = append(deletedKeys,
				&stateChange.DataTrieChange{
					Type:      stateChange.Write,
					Key:       []byte(key),
					Val:       originalVal,
					Version:   0,
					Operation: uint32(stateChange.Delete),
				},
			)
		}

		dataTrieKey, err := tdt.modifyTrie([]byte(key), dataEntry, oldVal, dtr)
		if err != nil {
			tdt.rollbackAppliedUpdates(dtr, trieUpdates)
			return nil, nil, err
		}

		index++

		isFirstMigration := oldVal.Version == core.NotSpecified && dataEntry.newVersion == core.AutoBalanceEnabled
		if isFirstMigration && len(dataTrieKey) != 0 {
			if !wasDeleted {
				trieUpdates = append(trieUpdates, oldVal)
			}

			deletedData := core.TrieData{
				Key:   dataTrieKey,
				Value: nil,
			}
			oldValues = append(oldValues, deletedData)
			trieUpdates = append(trieUpdates, deletedData)
		} else if len(dataTrieKey) != 0 && !wasDeleted {
			trieUpdates = append(trieUpdates, oldVal)
		}

		if len(dataTrieKey) == 0 {
			continue
		}

		if dataEntry.index > len(newData)-1 {
			tdt.rollbackAppliedUpdates(dtr, trieUpdates)
			return nil, nil, fmt.Errorf("index out of range")
		}

		dataTrieOperation := uint32(stateChange.NotSpecified)
		version := dataEntry.newVersion
		val := dataEntry.value
		if len(val) == 0 {
			dataTrieOperation = uint32(stateChange.Delete)
			version = oldVal.Version
			val, err = tdt.getValueWithoutMetadata([]byte(key), oldVal)
			if err != nil {
				tdt.rollbackAppliedUpdates(dtr, trieUpdates)
				return nil, nil, err
			}
		}

		newData[dataEntry.index] = &stateChange.DataTrieChange{
			Type:      stateChange.Write,
			Key:       []byte(key),
			Val:       val,
			Version:   uint32(version),
			Operation: dataTrieOperation,
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

func (tdt *trackableDataTrie) collectReadOperation(key []byte, val []byte, version uint32) {
	sc := &stateChange.StateAccess{
		Type:        stateChange.Read,
		MainTrieKey: tdt.identifier,
		MainTrieVal: nil,
		Operation:   stateChange.GetDataTrieValue,
		DataTrieChanges: []*stateChange.DataTrieChange{
			{
				Type:    stateChange.Read,
				Key:     key,
				Val:     val,
				Version: version,
			},
		},
	}
	tdt.stateAccessesCollector.AddStateAccess(sc)
}

func (tdt *trackableDataTrie) retrieveValueFromTrie(key []byte, writesHashedKey bool) (core.TrieData, uint32, error) {
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindData)
	defer stateAnalysis.SetTrieKind(prevKind)

	legacyAtHashedKey := false
	if tdt.enableEpochsHandler.IsFlagEnabled(common.AutoBalanceDataTriesFlag) {
		hashedKey := tdt.hasher.Compute(string(key))
		stateAnalysis.Inc(stateAnalysis.CDtKeyHashComputed)
		stateAnalysis.Inc(stateAnalysis.CDtLookupHashed)
		valWithMetadata, depth, version, err := tdt.tr.GetWithVersion(hashedKey)
		if err != nil {
			return core.TrieData{}, 0, err
		}
		if len(valWithMetadata) != 0 {
			stateAnalysis.Inc(stateAnalysis.CDtLookupHashedHit)
			stateAnalysis.TouchDataTrieKey(tdt.identifier, key, len(valWithMetadata))
		}

		versionMismatch := len(valWithMetadata) != 0 && version != uint32(core.AutoBalanceEnabled)
		validateVersion := versionMismatch && tdt.enableEpochsHandler.IsFlagEnabled(common.FixDataTrieLeafValidationFlag)
		reportedVersion := uint32(core.AutoBalanceEnabled)
		if validateVersion {
			reportedVersion = version
		}
		tdt.collectReadOperation(hashedKey, valWithMetadata, reportedVersion)

		if len(valWithMetadata) != 0 {
			if validateVersion {
				if version != uint32(core.NotSpecified) || writesHashedKey {
					return core.TrieData{}, depth, state.ErrInvalidDataTrieLeaf
				}
				// Reads and deletions may use the raw key without overwriting this legacy leaf.
				legacyAtHashedKey = true
			} else {
				trieValue := core.TrieData{
					Key:     hashedKey,
					Value:   valWithMetadata,
					Version: core.AutoBalanceEnabled,
				}

				return trieValue, depth, nil
			}
		}
	}

	stateAnalysis.Inc(stateAnalysis.CDtLookupRaw)
	valWithMetadata, depth, version, err := tdt.tr.GetWithVersion(key)
	if err != nil {
		return core.TrieData{}, 0, err
	}
	if len(valWithMetadata) != 0 {
		stateAnalysis.Inc(stateAnalysis.CDtLookupRawHit)
		stateAnalysis.TouchDataTrieKey(tdt.identifier, key, len(valWithMetadata))
	} else {
		stateAnalysis.Inc(stateAnalysis.CDtLookupMiss)
		stateAnalysis.TouchDataTrieKey(tdt.identifier, key, 0)
	}
	versionMismatch := len(valWithMetadata) != 0 && version != uint32(core.NotSpecified)
	validateVersion := versionMismatch && tdt.enableEpochsHandler.IsFlagEnabled(common.FixDataTrieLeafValidationFlag)
	reportedVersion := uint32(core.NotSpecified)
	if validateVersion {
		reportedVersion = version
	}
	tdt.collectReadOperation(key, valWithMetadata, reportedVersion)
	if len(valWithMetadata) != 0 {
		if validateVersion {
			return core.TrieData{}, depth, state.ErrInvalidDataTrieLeaf
		}
		trieValue := core.TrieData{
			Key:     key,
			Value:   valWithMetadata,
			Version: core.NotSpecified,
		}

		return trieValue, depth, nil
	}

	if legacyAtHashedKey {
		// Journal the missing raw key, not the occupied hashed key, when reverting a no-op deletion.
		return core.TrieData{Key: key, Version: core.NotSpecified}, depth, nil
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
		return tdt.getValueAutoBalanceVersion(key, trieData.Value)
	}

	return tdt.getValueNotSpecifiedVersion(key, trieData.Value), nil
}

func (tdt *trackableDataTrie) getValueAutoBalanceVersion(key, val []byte) ([]byte, error) {
	dataTrieVal := &dataTrieValue.TrieLeafData{}
	err := tdt.marshaller.Unmarshal(dataTrieVal, val)
	if err != nil {
		return nil, err
	}

	validBinding := bytes.Equal(dataTrieVal.Key, key) && bytes.Equal(dataTrieVal.Address, tdt.identifier)
	if !validBinding && tdt.enableEpochsHandler.IsFlagEnabled(common.FixDataTrieLeafValidationFlag) {
		return nil, state.ErrInvalidDataTrieLeaf
	}
	return dataTrieVal.Value, nil
}

func (tdt *trackableDataTrie) getValueNotSpecifiedVersion(key []byte, val []byte) []byte {
	tailLength := len(key) + len(tdt.identifier)
	trimmedValue, _ := common.TrimSuffixFromValue(val, tailLength)

	return trimmedValue
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

func (tdt *trackableDataTrie) modifyTrie(key []byte, dataEntry dirtyData, oldVal core.TrieData, dtr state.DataTrie) ([]byte, error) {
	prevKind := stateAnalysis.SetTrieKind(stateAnalysis.KindData)
	defer stateAnalysis.SetTrieKind(prevKind)

	version := dataEntry.newVersion
	newKey := tdt.getKeyForVersion(key, version)

	if len(dataEntry.value) == 0 {
		return tdt.deleteFromTrie(oldVal, key, dtr)
	}

	value, err := tdt.getValueForVersion(key, dataEntry.value, version)
	if err != nil {
		return nil, err
	}

	err = dtr.UpdateWithVersion(newKey, value, version)
	if err != nil {
		return nil, err
	}

	return newKey, nil
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

// classifyUpdate records what kind of change a dirty key produces on the trie, by comparing it with the old value
func (tdt *trackableDataTrie) classifyUpdate(key []byte, dataEntry dirtyData, oldVal core.TrieData) {
	if !stateAnalysis.Enabled() {
		return
	}

	stateAnalysis.Add(stateAnalysis.CDtUpdateBytesNew, int64(len(dataEntry.value)))
	stateAnalysis.Add(stateAnalysis.CDtUpdateBytesOld, int64(len(oldVal.Value)))

	if len(dataEntry.value) == 0 {
		if len(oldVal.Value) == 0 {
			stateAnalysis.Inc(stateAnalysis.CDtUpdateDeleteMiss)
			return
		}
		stateAnalysis.Inc(stateAnalysis.CDtUpdateDelete)
		return
	}
	if len(oldVal.Value) == 0 {
		stateAnalysis.Inc(stateAnalysis.CDtUpdateInsert)
		return
	}
	if oldVal.Version != dataEntry.newVersion {
		stateAnalysis.Inc(stateAnalysis.CDtUpdateMigrate)
		return
	}

	oldValue, err := tdt.getValueWithoutMetadata(key, oldVal)
	if err == nil && bytes.Equal(oldValue, dataEntry.value) {
		stateAnalysis.Inc(stateAnalysis.CDtUpdateNoop)
		return
	}
	stateAnalysis.Inc(stateAnalysis.CDtUpdateModify)
}
