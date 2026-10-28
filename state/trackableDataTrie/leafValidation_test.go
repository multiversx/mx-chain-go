package trackableDataTrie_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
)

func TestLeafValidation_ReadAndWrite(t *testing.T) {
	for _, scenario := range []string{"valid", "legacy alias", "wrong key", "wrong address", "malformed", "unknown version", "modern at raw key"} {
		for _, active := range []bool{false, true} {
			t.Run(scenario+map[bool]string{false: "/before", true: "/after"}[active], func(t *testing.T) {
				hasher := blake2b.NewBlake2b()
				codec := &marshal.GogoProtoMarshalizer{}
				address := bytes.Repeat([]byte{7}, 32)
				key := []byte("ELRONDesdtEXAMPLE-123456")
				physical := hasher.Compute(string(key))
				version := core.AutoBalanceEnabled
				decoded := &dataTrieValue.TrieLeafData{Key: key, Address: address, Value: []byte{42}}
				if scenario == "wrong key" {
					decoded.Key = []byte("other")
				}
				if scenario == "wrong address" {
					decoded.Address = []byte("other")
				}
				value, err := codec.Marshal(decoded)
				require.NoError(t, err)
				switch scenario {
				case "legacy alias":
					version = core.NotSpecified
					value = append(append([]byte{0x0a, 1, 42, 0x1a, 64}, physical...), address...)
				case "malformed":
					value = []byte{0xff}
				case "unknown version":
					version = core.TrieNodeVersion(99)
				case "modern at raw key":
					physical = key
				}
				flagChecks := 0
				flags := &enableEpochsHandlerMock.EnableEpochsHandlerStub{IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					if flag == common.FixDataTrieLeafValidationFlag {
						flagChecks++
						return active
					}
					return flag == common.AutoBalanceDataTriesFlag
				}}
				tr, err := trie.NewTrie(trie.TrieArgs{TrieStorage: &storageManager.StorageManagerStub{}, Marshalizer: codec, Hasher: hasher, EnableEpochsHandler: flags, MaxTrieLevelInMemory: 10, Throttler: trie.NewDisabledTrieGoRoutinesThrottler()})
				require.NoError(t, err)
				tr.UpdateWithVersion(physical, value, version)
				before, err := tr.RootHash()
				require.NoError(t, err)
				tdt, err := trackableDataTrie.NewTrackableDataTrie(address, hasher, codec, flags, &stateMock.StateAccessesCollectorStub{})
				require.NoError(t, err)
				tdt.SetDataTrie(tr)
				got, _, err := tdt.RetrieveValue(key)
				if scenario == "malformed" || (active && scenario != "valid" && scenario != "legacy alias") {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					if active && scenario == "legacy alias" {
						require.Empty(t, got)
					} else if scenario != "modern at raw key" {
						require.Equal(t, []byte{42}, got)
					}
				}
				require.NoError(t, tdt.SaveKeyValue(key, []byte("updated")))
				_, _, err = tdt.SaveDirtyData(tr)
				if active && scenario != "valid" {
					require.Error(t, err)
					after, rootErr := tr.RootHash()
					require.NoError(t, rootErr)
					require.Equal(t, before, after)
				} else {
					require.NoError(t, err)
				}
				if scenario == "valid" {
					require.Zero(t, flagChecks, "normal reads and writes must not check activation")
				}
			})
		}
	}
}

func TestLeafValidation_LegacyMigrationAndRevert(t *testing.T) {
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AutoBalanceDataTriesFlag, common.FixDataTrieLeafValidationFlag)
	hasher := blake2b.NewBlake2b()
	codec := &marshal.GogoProtoMarshalizer{}
	tr, err := trie.NewTrie(trie.TrieArgs{TrieStorage: &storageManager.StorageManagerStub{}, Marshalizer: codec, Hasher: hasher, EnableEpochsHandler: flags, MaxTrieLevelInMemory: 10, Throttler: trie.NewDisabledTrieGoRoutinesThrottler()})
	require.NoError(t, err)
	address, key, value := []byte("account"), []byte("legacy-key"), []byte("old-value")
	stored := append(append(bytes.Clone(value), key...), address...)
	tr.UpdateWithVersion(key, stored, core.NotSpecified)
	before, err := tr.RootHash()
	require.NoError(t, err)
	tdt, err := trackableDataTrie.NewTrackableDataTrie(address, hasher, codec, flags, &stateMock.StateAccessesCollectorStub{})
	require.NoError(t, err)
	tdt.SetDataTrie(tr)
	got, _, err := tdt.RetrieveValue(key)
	require.NoError(t, err)
	require.Equal(t, value, got)
	require.NoError(t, tdt.SaveKeyValue(key, []byte("updated")))
	_, oldValues, err := tdt.SaveDirtyData(tr)
	require.NoError(t, err)
	got, _, err = tr.Get(key)
	require.NoError(t, err)
	require.Empty(t, got)
	_, _, version, err := tr.GetWithVersion(hasher.Compute(string(key)))
	require.NoError(t, err)
	require.Equal(t, uint32(core.AutoBalanceEnabled), version)
	for i := len(oldValues) - 1; i >= 0; i-- {
		old := oldValues[i]
		tr.UpdateWithVersion(old.Key, old.Value, old.Version)
	}
	after, err := tr.RootHash()
	require.NoError(t, err)
	require.Equal(t, before, after)
	got, _, err = tdt.RetrieveValue(key)
	require.NoError(t, err)
	require.Equal(t, value, got)
	_, _, err = tdt.RetrieveValue([]byte("absent"))
	require.NoError(t, err)
	require.NoError(t, tdt.SaveKeyValue(key, nil))
	_, _, err = tdt.SaveDirtyData(tr)
	require.NoError(t, err)
	got, _, err = tr.Get(key)
	require.NoError(t, err)
	require.Empty(t, got)
	var _ state.DataTrie = tr
}

type validationFailureAfterWrite struct {
	state.DataTrie
	writes       int
	invalidValue []byte
}

func (tr *validationFailureAfterWrite) UpdateWithVersion(key, value []byte, version core.TrieNodeVersion) {
	tr.writes++
	tr.DataTrie.UpdateWithVersion(key, value, version)
}

func (tr *validationFailureAfterWrite) GetWithVersion(key []byte) ([]byte, uint32, uint32, error) {
	value, depth, version, err := tr.DataTrie.GetWithVersion(key)
	if tr.writes > 0 && err == nil && len(value) > 0 {
		return tr.invalidValue, depth, version, nil
	}
	return value, depth, version, err
}

func TestLeafValidation_RollsBackEarlierWritesOnBindingFailure(t *testing.T) {
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AutoBalanceDataTriesFlag, common.FixDataTrieLeafValidationFlag)
	hasher := blake2b.NewBlake2b()
	codec := &marshal.GogoProtoMarshalizer{}
	tr, err := trie.NewTrie(trie.TrieArgs{TrieStorage: &storageManager.StorageManagerStub{}, Marshalizer: codec, Hasher: hasher, EnableEpochsHandler: flags, MaxTrieLevelInMemory: 10, Throttler: trie.NewDisabledTrieGoRoutinesThrottler()})
	require.NoError(t, err)
	address := []byte("account")
	for _, key := range []string{"first", "second"} {
		value, marshalErr := codec.Marshal(&dataTrieValue.TrieLeafData{Key: []byte(key), Address: address, Value: []byte("old")})
		require.NoError(t, marshalErr)
		tr.UpdateWithVersion(hasher.Compute(key), value, core.AutoBalanceEnabled)
	}
	before, err := tr.RootHash()
	require.NoError(t, err)
	invalidValue, err := codec.Marshal(&dataTrieValue.TrieLeafData{Key: []byte("wrong"), Address: address, Value: []byte("old")})
	require.NoError(t, err)
	wrapper := &validationFailureAfterWrite{DataTrie: tr, invalidValue: invalidValue}
	tdt, err := trackableDataTrie.NewTrackableDataTrie(address, hasher, codec, flags, &stateMock.StateAccessesCollectorStub{})
	require.NoError(t, err)
	tdt.SetDataTrie(wrapper)
	require.NoError(t, tdt.SaveKeyValue([]byte("first"), []byte("new")))
	require.NoError(t, tdt.SaveKeyValue([]byte("second"), []byte("new")))
	_, _, err = tdt.SaveDirtyData(tr)
	require.ErrorIs(t, err, state.ErrInvalidDataTrieLeaf)
	require.Equal(t, 2, wrapper.writes, "one write followed by its rollback")
	after, err := tr.RootHash()
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestLeafValidation_MigrationUsesLogicalKey(t *testing.T) {
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AutoBalanceDataTriesFlag, common.FixDataTrieLeafValidationFlag)
	hasher := blake2b.NewBlake2b()
	codec := &marshal.GogoProtoMarshalizer{}
	tr, err := trie.NewTrie(trie.TrieArgs{TrieStorage: &storageManager.StorageManagerStub{}, Marshalizer: codec, Hasher: hasher, EnableEpochsHandler: flags, MaxTrieLevelInMemory: 10, Throttler: trie.NewDisabledTrieGoRoutinesThrottler()})
	require.NoError(t, err)
	key, address := []byte("key"), []byte("account")
	value, err := codec.Marshal(&dataTrieValue.TrieLeafData{Key: key, Address: address, Value: []byte("value")})
	require.NoError(t, err)
	physical := hasher.Compute(string(key))
	tr.UpdateWithVersion(physical, value, core.AutoBalanceEnabled)
	tdt, err := trackableDataTrie.NewTrackableDataTrie(address, hasher, codec, flags, &stateMock.StateAccessesCollectorStub{})
	require.NoError(t, err)
	tdt.SetDataTrie(tr)
	err = tdt.MigrateDataTrieLeaves(vmcommon.ArgsMigrateDataTrieLeaves{
		OldVersion: core.AutoBalanceEnabled, NewVersion: core.AutoBalanceEnabled,
		TrieMigrator: &trieMock.DataTrieMigratorStub{GetLeavesToBeMigratedCalled: func() []core.TrieData {
			return []core.TrieData{{Key: physical, Value: value, Version: core.AutoBalanceEnabled}}
		}},
	})
	require.NoError(t, err)
	_, _, err = tdt.SaveDirtyData(tr)
	require.NoError(t, err)
	got, _, err := tdt.RetrieveValue(key)
	require.NoError(t, err)
	require.Equal(t, []byte("value"), got)
}

func TestLeafValidation_LegacyAliasFallback(t *testing.T) {
	for _, rawPresent := range []bool{false, true} {
		for _, operation := range []string{"read", "write", "delete"} {
			t.Run(map[bool]string{false: "absent/", true: "present/"}[rawPresent]+operation, func(t *testing.T) {
				flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AutoBalanceDataTriesFlag, common.FixDataTrieLeafValidationFlag)
				hasher := blake2b.NewBlake2b()
				codec := &marshal.GogoProtoMarshalizer{}
				tr, err := trie.NewTrie(trie.TrieArgs{TrieStorage: &storageManager.StorageManagerStub{}, Marshalizer: codec, Hasher: hasher, EnableEpochsHandler: flags, MaxTrieLevelInMemory: 10, Throttler: trie.NewDisabledTrieGoRoutinesThrottler()})
				require.NoError(t, err)
				key, address := []byte("ELRONDesdtEXAMPLE-123456"), bytes.Repeat([]byte{7}, 32)
				hashed := hasher.Compute(string(key))
				conflict := append(append([]byte{0x0a, 1, 42, 0x1a, 64}, hashed...), address...)
				tr.UpdateWithVersion(hashed, conflict, core.NotSpecified)
				var expected []byte
				if rawPresent {
					expected = []byte("legitimate value")
					stored := append(append(bytes.Clone(expected), key...), address...)
					tr.UpdateWithVersion(key, stored, core.NotSpecified)
				}
				before, err := tr.RootHash()
				require.NoError(t, err)
				tdt, err := trackableDataTrie.NewTrackableDataTrie(address, hasher, codec, flags, &stateMock.StateAccessesCollectorStub{})
				require.NoError(t, err)
				tdt.SetDataTrie(tr)
				switch operation {
				case "read":
					got, depth, readErr := tdt.RetrieveValue(key)
					require.NoError(t, readErr)
					require.Equal(t, expected, got)
					_, rawDepth, rawErr := tr.Get(key)
					require.NoError(t, rawErr)
					require.Equal(t, rawDepth, depth)
				case "write":
					require.NoError(t, tdt.SaveKeyValue(key, []byte("new value")))
					_, _, saveErr := tdt.SaveDirtyData(tr)
					require.ErrorIs(t, saveErr, state.ErrInvalidDataTrieLeaf)
				case "delete":
					require.NoError(t, tdt.SaveKeyValue(key, nil))
					_, oldValues, saveErr := tdt.SaveDirtyData(tr)
					require.NoError(t, saveErr)
					got, _, readErr := tr.Get(key)
					require.NoError(t, readErr)
					require.Empty(t, got)
					for i := len(oldValues) - 1; i >= 0; i-- {
						old := oldValues[i]
						tr.UpdateWithVersion(old.Key, old.Value, old.Version)
					}
				}
				gotConflict, _, version, err := tr.GetWithVersion(hashed)
				require.NoError(t, err)
				require.Equal(t, conflict, gotConflict)
				require.Equal(t, uint32(core.NotSpecified), version)
				after, err := tr.RootHash()
				require.NoError(t, err)
				require.Equal(t, before, after)
			})
		}
	}
}

func TestLeafValidation_InteractingBatchDeterministic(t *testing.T) {
	for _, aliasFirst := range []bool{false, true} {
		for _, deleteAlias := range []bool{false, true} {
			t.Run(fmt.Sprintf("aliasSortsFirst=%t/delete=%t", aliasFirst, deleteAlias), func(t *testing.T) {
				hasher := blake2b.NewBlake2b()
				var key, alias []byte
				for i := 0; i < 1000; i++ {
					key = []byte(fmt.Sprintf("key-%d", i))
					alias = hasher.Compute(string(key))
					if (bytes.Compare(alias, key) < 0) == aliasFirst {
						break
					}
				}
				require.Equal(t, aliasFirst, bytes.Compare(alias, key) < 0)
				var expectedRoot []byte
				for iteration := 0; iteration < 64; iteration++ {
					codec := &marshal.GogoProtoMarshalizer{}
					flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AutoBalanceDataTriesFlag, common.FixDataTrieLeafValidationFlag)
					tr, err := trie.NewTrie(trie.TrieArgs{TrieStorage: &storageManager.StorageManagerStub{}, Marshalizer: codec, Hasher: hasher, EnableEpochsHandler: flags, MaxTrieLevelInMemory: 10, Throttler: trie.NewDisabledTrieGoRoutinesThrottler()})
					require.NoError(t, err)
					address := []byte("account")
					legacy := append(append([]byte("old"), alias...), address...)
					tr.UpdateWithVersion(alias, legacy, core.NotSpecified)
					before, err := tr.RootHash()
					require.NoError(t, err)
					tdt, err := trackableDataTrie.NewTrackableDataTrie(address, hasher, codec, flags, &stateMock.StateAccessesCollectorStub{})
					require.NoError(t, err)
					tdt.SetDataTrie(tr)
					aliasValue := []byte("updated alias")
					if deleteAlias {
						aliasValue = nil
					}
					if iteration%2 == 0 {
						require.NoError(t, tdt.SaveKeyValue(alias, aliasValue))
						require.NoError(t, tdt.SaveKeyValue(key, []byte("new P")))
					} else {
						require.NoError(t, tdt.SaveKeyValue(key, []byte("new P")))
						require.NoError(t, tdt.SaveKeyValue(alias, aliasValue))
					}
					_, oldValues, err := tdt.SaveDirtyData(tr)
					if aliasFirst {
						require.NoError(t, err)
					} else {
						require.ErrorIs(t, err, state.ErrInvalidDataTrieLeaf)
					}
					after, err := tr.RootHash()
					require.NoError(t, err)
					if iteration == 0 {
						expectedRoot = after
					}
					require.Equal(t, expectedRoot, after)
					for i := len(oldValues) - 1; i >= 0; i-- {
						old := oldValues[i]
						tr.UpdateWithVersion(old.Key, old.Value, old.Version)
					}
					reverted, err := tr.RootHash()
					require.NoError(t, err)
					require.Equal(t, before, reverted)
				}
			})
		}
	}
}

func TestLeafValidation_CollectedVersions(t *testing.T) {
	for _, active := range []bool{false, true} {
		for _, scenario := range []string{"legacy alias", "unknown hashed", "modern hashed", "modern raw"} {
			t.Run(fmt.Sprintf("%s/active=%t", scenario, active), func(t *testing.T) {
				flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AutoBalanceDataTriesFlag)
				if active {
					flags.AddActiveFlags(common.FixDataTrieLeafValidationFlag)
				}
				hasher, codec := blake2b.NewBlake2b(), &marshal.GogoProtoMarshalizer{}
				tr, err := trie.NewTrie(trie.TrieArgs{TrieStorage: &storageManager.StorageManagerStub{}, Marshalizer: codec, Hasher: hasher, EnableEpochsHandler: flags, MaxTrieLevelInMemory: 10, Throttler: trie.NewDisabledTrieGoRoutinesThrottler()})
				require.NoError(t, err)
				key, address := []byte("key"), []byte("account")
				physical, version := hasher.Compute(string(key)), core.AutoBalanceEnabled
				value, err := codec.Marshal(&dataTrieValue.TrieLeafData{Key: key, Address: address, Value: []byte("value")})
				require.NoError(t, err)
				switch scenario {
				case "legacy alias":
					version = core.NotSpecified
				case "unknown hashed":
					version = core.TrieNodeVersion(99)
				case "modern raw":
					physical = key
				}
				tr.UpdateWithVersion(physical, value, version)
				var reads []*stateChange.DataTrieChange
				collector := &stateMock.StateAccessesCollectorStub{AddStateChangeCalled: func(access *stateChange.StateAccess) {
					reads = append(reads, access.DataTrieChanges...)
				}}
				tdt, err := trackableDataTrie.NewTrackableDataTrie(address, hasher, codec, flags, collector)
				require.NoError(t, err)
				tdt.SetDataTrie(tr)
				_, _, _ = tdt.RetrieveValue(key)
				found := false
				for _, read := range reads {
					if !bytes.Equal(read.Key, physical) {
						continue
					}
					found = true
					expected := uint32(version)
					if !active {
						expected = uint32(core.AutoBalanceEnabled)
						if scenario == "modern raw" {
							expected = uint32(core.NotSpecified)
						}
					}
					require.Equal(t, expected, read.Version)
					require.Equal(t, value, read.Val)
				}
				require.True(t, found)
			})
		}
	}
}
