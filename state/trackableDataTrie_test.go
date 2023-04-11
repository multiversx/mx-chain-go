package state_test

import (
	"bytes"
	"fmt"
	"github.com/multiversx/mx-chain-go/state/trieValuesCache"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateTest "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackableDataTrie(t *testing.T) {
	t.Parallel()

	t.Run("create with nil hasher", func(t *testing.T) {
		t.Parallel()

		tdt, err := state.NewTrackableDataTrie(
			[]byte("identifier"),
			&trieMock.TrieStub{},
			nil,
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		assert.Equal(t, state.ErrNilHasher, err)
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("create with nil marshaller", func(t *testing.T) {
		t.Parallel()

		tdt, err := state.NewTrackableDataTrie(
			[]byte("identifier"),
			&trieMock.TrieStub{},
			&hashingMocks.HasherMock{},
			nil,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		assert.Equal(t, state.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("create with nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		tdt, err := state.NewTrackableDataTrie(
			[]byte("identifier"),
			&trieMock.TrieStub{},
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			nil,
			&stateTest.TrieValuesCacherStub{},
		)
		assert.Equal(t, state.ErrNilEnableEpochsHandler, err)
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("create with nil trieValuesCacher", func(t *testing.T) {
		t.Parallel()

		tdt, err := state.NewTrackableDataTrie(
			[]byte("identifier"),
			&trieMock.TrieStub{},
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			nil,
		)
		assert.Equal(t, state.ErrNilTrieValuesCacher, err)
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tdt, err := state.NewTrackableDataTrie(
			[]byte("identifier"),
			&trieMock.TrieStub{},
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(tdt))
	})
}

func TestTrackableDataTrie_SaveKeyValue(t *testing.T) {
	t.Parallel()

	t.Run("data too large", func(t *testing.T) {
		t.Parallel()

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			&trieMock.TrieStub{},
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)

		err := tdt.SaveKeyValue([]byte("key"), make([]byte, core.MaxLeafSize+1))
		assert.Equal(t, err, data.ErrLeafSizeTooBig)
	})

	t.Run("should save given val only in dirty data", func(t *testing.T) {
		t.Parallel()

		keyExpected := []byte("key")
		value := []byte("value")
		trie := &trieMock.TrieStub{
			UpdateCalled: func(key, value []byte) error {
				assert.Fail(t, "should not have saved directly in the trie")
				return nil
			},
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				assert.Fail(t, "should not have saved directly in the trie")
				return nil, 0, nil
			},
		}
		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		assert.NotNil(t, tdt)

		_ = tdt.SaveKeyValue(keyExpected, value)

		dirtyData := tdt.DirtyData()
		assert.Equal(t, 1, len(dirtyData))
		assert.Equal(t, value, dirtyData[string(keyExpected)].Value)
	})
}

func TestTrackableDataTrie_RetrieveValue(t *testing.T) {
	t.Parallel()

	t.Run("should check dirty data first", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("id")
		key := []byte("key")
		tail := append(key, identifier...)
		retrievedTrieVal := []byte("value")
		trieValue := append(retrievedTrieVal, tail...)
		newTrieValue := []byte("new trie value")
		numTvcCalls := 0

		trie := &trieMock.TrieStub{
			GetCalled: func(trieKey []byte) ([]byte, uint32, error) {
				if bytes.Equal(trieKey, key) {
					return trieValue, 0, nil
				}
				return nil, 0, nil
			},
		}
		tvc := &stateTest.TrieValuesCacherStub{
			GetCalled: func(key []byte) (core.TrieData, bool) {
				numTvcCalls++
				return core.TrieData{}, false
			},
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			tvc,
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(key)
		assert.Equal(t, retrievedTrieVal, valRecovered)
		assert.Nil(t, err)
		assert.Equal(t, 1, numTvcCalls)

		_ = tdt.SaveKeyValue(key, newTrieValue)
		valRecovered, _, err = tdt.RetrieveValue(key)
		assert.Equal(t, newTrieValue, valRecovered)
		assert.Nil(t, err)
		assert.Equal(t, 1, numTvcCalls)
	})

	t.Run("should verify trieValuesCache before trie", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		key := []byte("key")
		val := []byte("value")
		valWithMetadata := append(val, key...)
		valWithMetadata = append(valWithMetadata, identifier...)
		trieData := core.TrieData{
			Value:   valWithMetadata,
			Version: core.NotSpecified,
		}

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				assert.Fail(t, "should not have called trie.Get")
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}

		tvc := &stateTest.TrieValuesCacherStub{
			GetCalled: func(key []byte) (core.TrieData, bool) {
				return trieData, true
			},
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			enableEpochsHandler,
			tvc,
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(key)
		assert.Nil(t, err)
		assert.Equal(t, val, valRecovered)
	})

	t.Run("nil data trie should err", func(t *testing.T) {
		t.Parallel()

		tdt, err := state.NewTrackableDataTrie(
			[]byte("identifier"),
			nil,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		assert.Nil(t, err)
		assert.NotNil(t, tdt)

		_, _, err = tdt.RetrieveValue([]byte("ABC"))
		assert.Equal(t, state.ErrNilTrie, err)
	})

	t.Run("val with appended data found in trie", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		expectedVal := []byte("value")
		valueWithMetadata := append(expectedVal, expectedKey...)
		valueWithMetadata = append(valueWithMetadata, identifier...)
		putInTvcCalled := false

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, expectedKey) {
					return valueWithMetadata, 0, nil
				}
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tvc := &stateTest.TrieValuesCacherStub{
			PutCalled: func(key []byte, value core.TrieData) {
				putInTvcCalled = true
				assert.Equal(t, expectedKey, key)
				assert.Equal(t, key, value.Key)
				assert.Equal(t, valueWithMetadata, value.Value)
				assert.Equal(t, core.NotSpecified, value.Version)

			},
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			enableEpochsHandler,
			tvc,
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, valRecovered)
		assert.True(t, putInTvcCalled)
	})

	t.Run("autoBalance data tries disabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		expectedVal := []byte("value")
		valueWithMetadata := append(expectedVal, expectedKey...)
		valueWithMetadata = append(valueWithMetadata, identifier...)
		hasher := &hashingMocks.HasherMock{}
		putInTvcCalled := false

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, expectedKey) {
					return valueWithMetadata, 0, nil
				}
				if bytes.Equal(key, hasher.Compute(string(expectedKey))) {
					assert.Fail(t, "this should not have been called")
				}
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: false,
		}
		tvc := &stateTest.TrieValuesCacherStub{
			PutCalled: func(key []byte, value core.TrieData) {
				putInTvcCalled = true
				assert.Equal(t, expectedKey, key)
				assert.Equal(t, key, value.Key)
				assert.Equal(t, valueWithMetadata, value.Value)
				assert.Equal(t, core.NotSpecified, value.Version)

			},
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			enableEpochsHandler,
			tvc,
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, valRecovered)
		assert.True(t, putInTvcCalled)
	})

	t.Run("val as struct found in trie", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		expectedVal := []byte("value")
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		leafData := &dataTrieValue.TrieLeafData{
			Value:   expectedVal,
			Key:     expectedKey,
			Address: identifier,
		}
		valueWithMetadata, _ := marshaller.Marshal(leafData)
		putInTvcCalled := false

		trie := &trieMock.TrieStub{
			UpdateCalled: func(key, value []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, hasher.Compute(string(expectedKey))) {
					return valueWithMetadata, 0, nil
				}
				return nil, 0, nil
			},
		}
		tvc := &stateTest.TrieValuesCacherStub{
			PutCalled: func(key []byte, value core.TrieData) {
				putInTvcCalled = true
				assert.Equal(t, expectedKey, key)
				assert.Equal(t, hasher.Compute(string(key)), value.Key)
				assert.Equal(t, valueWithMetadata, value.Value)
				assert.Equal(t, core.AutoBalanceEnabled, value.Version)
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, valRecovered)
		assert.True(t, putInTvcCalled)
	})

	t.Run("trie malfunction should err", func(t *testing.T) {
		t.Parallel()

		errExpected := errors.New("expected err")
		keyExpected := []byte("key")
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, errExpected
			},
		}
		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(keyExpected)
		assert.Equal(t, errExpected, err)
		assert.Nil(t, valRecovered)
	})

	t.Run("val not found in trie - auto balance enabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		putInTvcCalled := false

		trie := &trieMock.TrieStub{
			UpdateCalled: func(key, value []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
		}
		tvc := &stateTest.TrieValuesCacherStub{
			PutCalled: func(key []byte, value core.TrieData) {
				putInTvcCalled = true
				assert.Equal(t, expectedKey, key)
				assert.Equal(t, hasher.Compute(string(key)), value.Key)
				assert.Equal(t, []byte(nil), value.Value)
				assert.Equal(t, core.AutoBalanceEnabled, value.Version)
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, []byte(nil), valRecovered)
		assert.True(t, putInTvcCalled)
	})

	t.Run("val not found in trie - auto balance disabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		putInTvcCalled := false

		trie := &trieMock.TrieStub{
			UpdateCalled: func(key, value []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
		}
		tvc := &stateTest.TrieValuesCacherStub{
			PutCalled: func(key []byte, value core.TrieData) {
				putInTvcCalled = true
				assert.Equal(t, expectedKey, key)
				assert.Equal(t, key, value.Key)
				assert.Equal(t, []byte(nil), value.Value)
				assert.Equal(t, core.NotSpecified, value.Version)
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: false,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)
		assert.NotNil(t, tdt)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, []byte(nil), valRecovered)
		assert.True(t, putInTvcCalled)
	})
}

func TestTrackableDataTrie_SaveDirtyData(t *testing.T) {
	t.Parallel()

	t.Run("no dirty data", func(t *testing.T) {
		t.Parallel()

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
		}

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			nil,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			tvc,
		)

		oldValues, err := tdt.SaveDirtyData(&trieMock.TrieStub{})
		assert.Nil(t, err)
		assert.Equal(t, 0, len(oldValues))
		assert.True(t, cleanCalled)
	})

	t.Run("nil trie creates a new trie", func(t *testing.T) {
		t.Parallel()

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
		}

		recreateCalled := false
		trie := &trieMock.TrieStub{
			RecreateCalled: func(root []byte) (common.Trie, error) {
				recreateCalled = true
				return &trieMock.TrieStub{
					GetCalled: func(_ []byte) ([]byte, uint32, error) {
						return nil, 0, nil
					},
					UpdateWithVersionCalled: func(_, _ []byte, _ core.TrieNodeVersion) error {
						return nil
					},
				}, nil
			},
		}
		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			nil,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			tvc,
		)

		key := []byte("key")
		_ = tdt.SaveKeyValue(key, []byte("val"))
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, key, oldValues[0].Key)
		assert.Equal(t, []byte(nil), oldValues[0].Value)
		assert.True(t, recreateCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("present in trie as valWithAppendedData", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		expectedVal := []byte("value")
		valueWithMetadata := append(expectedVal, expectedKey...)
		valueWithMetadata = append(valueWithMetadata, identifier...)
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		deleteCalled := false
		updateCalled := false

		trieVal := &dataTrieValue.TrieLeafData{
			Value:   expectedVal,
			Key:     expectedKey,
			Address: identifier,
		}
		serializedTrieVal, _ := marshaller.Marshal(trieVal)

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, expectedKey) {
					return valueWithMetadata, 0, nil
				}
				return nil, 0, nil
			},
			UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
				assert.Equal(t, hasher.Compute(string(expectedKey)), key)
				assert.Equal(t, serializedTrieVal, value)
				updateCalled = true
				return nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, expectedKey, key)
				deleteCalled = true
				return nil
			},
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, expectedVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, expectedKey, oldValues[0].Key)
		assert.Equal(t, valueWithMetadata, oldValues[0].Value)
		assert.True(t, deleteCalled)
		assert.True(t, updateCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("present in trie as valWithAppendedData and auto balancing disabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		val := []byte("value")
		expectedVal := append(val, expectedKey...)
		expectedVal = append(expectedVal, identifier...)
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		updateCalled := false

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, expectedKey) {
					return expectedVal, 0, nil
				}
				return nil, 0, nil
			},
			UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
				assert.Equal(t, expectedKey, key)
				assert.Equal(t, expectedVal, value)
				updateCalled = true
				return nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Fail(t, "this should not have been called")
				return nil
			},
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: false,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, val)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, expectedKey, oldValues[0].Key)
		assert.Equal(t, expectedVal, oldValues[0].Value)
		assert.True(t, updateCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("present in trie as valAsStruct", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		newVal := []byte("value")
		oldVal := []byte("old val")
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		updateCalled := false

		oldTrieVal := &dataTrieValue.TrieLeafData{
			Value:   oldVal,
			Key:     expectedKey,
			Address: identifier,
		}
		serializedOldTrieVal, _ := marshaller.Marshal(oldTrieVal)

		newTrieVal := &dataTrieValue.TrieLeafData{
			Value:   newVal,
			Key:     expectedKey,
			Address: identifier,
		}
		serializedNewTrieVal, _ := marshaller.Marshal(newTrieVal)

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, hasher.Compute(string(expectedKey))) {
					return serializedOldTrieVal, 0, nil
				}
				return nil, 0, nil
			},
			UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
				assert.Equal(t, hasher.Compute(string(expectedKey)), key)
				assert.Equal(t, serializedNewTrieVal, value)
				updateCalled = true
				return nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Fail(t, "this delete should not have been called")
				return nil
			},
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, newVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, hasher.Compute(string(expectedKey)), oldValues[0].Key)
		assert.Equal(t, serializedOldTrieVal, oldValues[0].Value)
		assert.True(t, updateCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("not present in trie - autobalance enabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		newVal := []byte("value")
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		updateCalled := false

		newTrieVal := &dataTrieValue.TrieLeafData{
			Value:   newVal,
			Key:     expectedKey,
			Address: identifier,
		}
		serializedNewTrieVal, _ := marshaller.Marshal(newTrieVal)

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
			UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
				assert.Equal(t, hasher.Compute(string(expectedKey)), key)
				assert.Equal(t, serializedNewTrieVal, value)
				updateCalled = true
				return nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Fail(t, "this delete should not have been called")
				return nil
			},
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, newVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, hasher.Compute(string(expectedKey)), oldValues[0].Key)
		assert.Equal(t, []byte(nil), oldValues[0].Value)
		assert.True(t, updateCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("not present in trie - autobalance disabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		newVal := []byte("value")
		valueWithMetadata := append(newVal, expectedKey...)
		valueWithMetadata = append(valueWithMetadata, identifier...)
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}
		updateCalled := false

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
			UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
				assert.Equal(t, expectedKey, key)
				assert.Equal(t, valueWithMetadata, value)
				updateCalled = true
				return nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Fail(t, "this delete should not have been called")
				return nil
			},
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: false,
		}
		tdt, _ := state.NewTrackableDataTrie(
			identifier,
			trie,
			hasher,
			marshaller,
			enableEpochsHandler,
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, newVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, expectedKey, oldValues[0].Key)
		assert.Equal(t, []byte(nil), oldValues[0].Value)
		assert.True(t, updateCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("dirty data is reset", func(t *testing.T) {
		t.Parallel()

		expectedKey := []byte("key")
		val := []byte("value")

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
			UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
				return nil
			},
		}

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
		}

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, val)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.True(t, cleanCalled)
	})

	t.Run("nil val autobalance disabled", func(t *testing.T) {
		t.Parallel()

		expectedKey := []byte("key")
		retrievedVal := []byte("value")
		updateCalled := false
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return retrievedVal, 0, nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, expectedKey, key)
				updateCalled = true
				return nil
			},
		}

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.True(t, updateCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("nil val and nil old val", func(t *testing.T) {
		t.Parallel()

		expectedKey := []byte("key")
		deleteCalled := false
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, expectedKey, key)
				deleteCalled = true
				return nil
			},
		}

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.False(t, deleteCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("nil val autobalance enabled, old val saved at hashedKey", func(t *testing.T) {
		t.Parallel()

		hasher := &hashingMocks.HasherMock{}
		expectedKey := []byte("key")
		retrievedVal := []byte("value")
		deleteCalled := false
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(hasher.Compute(string(expectedKey)), key) {
					return retrievedVal, 0, nil
				}

				return nil, 0, nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, hasher.Compute(string(expectedKey)), key)
				deleteCalled = true
				return nil
			},
		}

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		enableEpchs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			enableEpchs,
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.True(t, deleteCalled)
		assert.True(t, cleanCalled)
	})

	t.Run("nil val autobalance enabled, old val saved at key", func(t *testing.T) {
		t.Parallel()

		expectedKey := []byte("key")
		deleteCalled := 0
		retrievedVal := []byte("value")
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(expectedKey, key) {
					return retrievedVal, 0, nil
				}

				return nil, 0, nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, expectedKey, key)
				deleteCalled++
				return nil
			},
		}

		cleanCalled := false
		tvc := &stateTest.TrieValuesCacherStub{
			CleanCalled: func() {
				cleanCalled = true
			},
			PutCalled: func(key []byte, value core.TrieData) {
				assert.Fail(t, "should not have called put")
			},
		}

		enableEpchs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			trie,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			enableEpchs,
			tvc,
		)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.Equal(t, 1, deleteCalled)
		assert.True(t, cleanCalled)
	})
}

func TestTrackableDataTrie_MigrateDataTrieLeaves(t *testing.T) {
	t.Parallel()

	t.Run("nil trie", func(t *testing.T) {
		t.Parallel()

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			nil,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		err := tdt.MigrateDataTrieLeaves(core.NotSpecified, core.AutoBalanceEnabled, &trieMock.DataTrieMigratorStub{})
		assert.Equal(t, state.ErrNilTrie, err)
	})

	t.Run("nil trie migrator", func(t *testing.T) {
		t.Parallel()

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			&trieMock.TrieStub{},
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		err := tdt.MigrateDataTrieLeaves(core.NotSpecified, core.AutoBalanceEnabled, nil)
		assert.Equal(t, errorsCommon.ErrNilTrieMigrator, err)
	})

	t.Run("CollectLeavesForMigrationFails", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		tr := &trieMock.TrieStub{
			CollectLeavesForMigrationCalled: func(oldVersion core.TrieNodeVersion, newVersion core.TrieNodeVersion, trieMigrator vmcommon.DataTrieMigrator) error {
				return expectedErr
			},
		}

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			tr,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&stateTest.TrieValuesCacherStub{},
		)
		err := tdt.MigrateDataTrieLeaves(core.NotSpecified, core.AutoBalanceEnabled, &trieMock.DataTrieMigratorStub{})
		assert.Equal(t, expectedErr, err)
	})

	t.Run("leaves that need to be migrated are added to dirty data", func(t *testing.T) {
		t.Parallel()

		leavesToBeMigrated := []core.TrieData{
			{
				Key:     []byte("key1"),
				Value:   []byte("value1"),
				Version: core.NotSpecified,
			},
			{
				Key:     []byte("key2"),
				Value:   []byte("value2"),
				Version: core.NotSpecified,
			},
			{
				Key:     []byte("key3"),
				Value:   []byte("value3"),
				Version: core.NotSpecified,
			},
		}
		tr := &trieMock.TrieStub{
			CollectLeavesForMigrationCalled: func(oldVersion core.TrieNodeVersion, newVersion core.TrieNodeVersion, trieMigrator vmcommon.DataTrieMigrator) error {
				return nil
			},
		}
		dtm := &trieMock.DataTrieMigratorStub{
			GetLeavesToBeMigratedCalled: func() []core.TrieData {
				return leavesToBeMigrated
			},
		}
		enableEpchs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tvc := &stateTest.TrieValuesCacherStub{
			PutCalled: func(key []byte, value core.TrieData) {
				for i := range leavesToBeMigrated {
					if bytes.Equal(key, leavesToBeMigrated[i].Key) {
						assert.Equal(t, leavesToBeMigrated[i].Value, value.Value)
						assert.Equal(t, leavesToBeMigrated[i].Version, value.Version)
						return
					}
				}

				assert.Fail(t, "key not found")
			},
		}

		tdt, _ := state.NewTrackableDataTrie(
			[]byte("identifier"),
			tr,
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			enableEpchs,
			tvc,
		)
		err := tdt.MigrateDataTrieLeaves(core.NotSpecified, 100, dtm)
		assert.Nil(t, err)

		dirtyData := tdt.DirtyData()
		assert.Equal(t, len(leavesToBeMigrated), len(dirtyData))
		for i := range leavesToBeMigrated {
			d := dirtyData[string(leavesToBeMigrated[i].Key)]
			assert.Equal(t, leavesToBeMigrated[i].Value, d.Value)
			assert.Equal(t, core.TrieNodeVersion(100), d.Version)
		}
	})
}

func TestTrackableDataTrie_SetAndGetDataTrie(t *testing.T) {
	t.Parallel()

	trie := &trieMock.TrieStub{}
	cleanCalled := false
	tvc := &stateTest.TrieValuesCacherStub{
		CleanCalled: func() {
			cleanCalled = true
		},
	}

	tdt, _ := state.NewTrackableDataTrie(
		[]byte("identifier"),
		trie,
		&hashingMocks.HasherMock{},
		&marshallerMock.MarshalizerMock{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		tvc,
	)

	newTrie := &trieMock.TrieStub{}
	tdt.SetDataTrie(newTrie)
	assert.True(t, cleanCalled)
	assert.Equal(t, newTrie, tdt.DataTrie())
}

func TestTrackableDataTrie_MigrateMoreLeavesThanCacheSize(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	trieValuesCacheSize := 10
	numLeavesForMigration := trieValuesCacheSize + 2
	leavesForMigration := make([]core.TrieData, numLeavesForMigration)

	keysForDeletion := make(map[string]struct{})
	keysForUpdate := make(map[string]struct{})
	updatedVals := make(map[string][]byte)

	for i := 0; i < numLeavesForMigration; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := []byte(fmt.Sprintf("value%d", i))

		leavesForMigration[i] = core.TrieData{
			Key:     key,
			Value:   val,
			Version: core.NotSpecified,
		}

		keysForDeletion[string(key)] = struct{}{}
		hashedKey := hasher.Compute(string(key))
		keysForUpdate[string(hashedKey)] = struct{}{}
		leafData := dataTrieValue.TrieLeafData{
			Value:   val,
			Key:     key,
			Address: identifier,
		}
		leafDataBytes, err := marshaller.Marshal(leafData)
		assert.Nil(t, err)
		updatedVals[string(hashedKey)] = leafDataBytes

	}
	numDeletedLeaves := 0
	numUpdatedLeaves := 0
	numTrieGet := 0

	tr := &trieMock.TrieStub{
		CollectLeavesForMigrationCalled: func(_ core.TrieNodeVersion, _ core.TrieNodeVersion, _ vmcommon.DataTrieMigrator) error {
			return nil
		},
		DeleteCalled: func(key []byte) error {
			_, ok := keysForDeletion[string(key)]
			assert.True(t, ok)
			delete(keysForDeletion, string(key))
			numDeletedLeaves++
			return nil
		},
		UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
			_, ok := keysForUpdate[string(key)]
			assert.True(t, ok)
			delete(keysForUpdate, string(key))

			expectedVal, ok := updatedVals[string(key)]
			assert.True(t, ok)
			assert.Equal(t, expectedVal, value)

			assert.Equal(t, core.AutoBalanceEnabled, version)
			numUpdatedLeaves++
			return nil
		},
		GetCalled: func(key []byte) ([]byte, uint32, error) {
			if bytes.Equal(leavesForMigration[numTrieGet+trieValuesCacheSize].Key, key) {
				val := leavesForMigration[numTrieGet+trieValuesCacheSize].Value
				numTrieGet++

				return val, 0, nil
			}

			return nil, 0, nil
		},
	}

	tvc, err := trieValuesCache.NewTrieValuesCache(trieValuesCacheSize)
	assert.Nil(t, err)
	enableEpchs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsAutoBalanceDataTriesEnabledField: true,
	}

	tdt, _ := state.NewTrackableDataTrie(
		identifier,
		tr,
		hasher,
		marshaller,
		enableEpchs,
		tvc,
	)

	dtm := &trieMock.DataTrieMigratorStub{
		GetLeavesToBeMigratedCalled: func() []core.TrieData {
			return leavesForMigration
		},
	}
	err = tdt.MigrateDataTrieLeaves(core.NotSpecified, core.AutoBalanceEnabled, dtm)
	assert.Nil(t, err)

	dirtyData := tdt.DirtyData()
	for i := 0; i < numLeavesForMigration; i++ {
		newDataEntry := dirtyData[string(leavesForMigration[i].Key)]
		assert.Equal(t, leavesForMigration[i].Value, newDataEntry.Value)
		assert.Equal(t, core.AutoBalanceEnabled, newDataEntry.Version)
	}

	for i := 0; i < trieValuesCacheSize; i++ {
		trieData, ok := tvc.Get(leavesForMigration[i].Key)
		assert.True(t, ok)
		assert.Equal(t, leavesForMigration[i].Key, trieData.Key)
		assert.Equal(t, leavesForMigration[i].Value, trieData.Value)
		assert.Equal(t, core.NotSpecified, trieData.Version)
	}

	for i := trieValuesCacheSize; i < numLeavesForMigration; i++ {
		_, ok := tvc.Get(leavesForMigration[i].Key)
		assert.False(t, ok)
	}

	oldValues, err := tdt.SaveDirtyData(tr)
	assert.Nil(t, err)
	assert.Equal(t, numLeavesForMigration, len(oldValues))
	assert.Equal(t, numLeavesForMigration, numDeletedLeaves)
	assert.Equal(t, numLeavesForMigration, numUpdatedLeaves)
	assert.Equal(t, numLeavesForMigration-trieValuesCacheSize, numTrieGet)
}
