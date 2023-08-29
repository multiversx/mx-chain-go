package trackableDataTrie_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackableDataTrie(t *testing.T) {
	t.Parallel()

	t.Run("create with nil hasher", func(t *testing.T) {
		t.Parallel()

		tdt, err := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), nil, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.Equal(t, state.ErrNilHasher, err)
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("create with nil marshaller", func(t *testing.T) {
		t.Parallel()

		tdt, err := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, nil, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.Equal(t, state.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("create with nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		tdt, err := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, nil)
		assert.Equal(t, state.ErrNilEnableEpochsHandler, err)
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("create with invalid enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		tdt, err := trackableDataTrie.NewTrackableDataTrie(
			[]byte("identifier"),
			&hashingMocks.HasherMock{},
			&marshallerMock.MarshalizerMock{},
			enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined())
		assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
		assert.True(t, check.IfNil(tdt))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tdt, err := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(tdt))
	})
}

func TestTrackableDataTrie_SaveKeyValue(t *testing.T) {
	t.Parallel()

	t.Run("data too large", func(t *testing.T) {
		t.Parallel()

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

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
		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

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

		trie := &trieMock.TrieStub{
			GetCalled: func(trieKey []byte) ([]byte, uint32, error) {
				if bytes.Equal(trieKey, key) {
					return trieValue, 0, nil
				}
				return nil, 0, nil
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

		valRecovered, _, err := tdt.RetrieveValue(key)
		assert.Equal(t, retrievedTrieVal, valRecovered)
		assert.Nil(t, err)

		_ = tdt.SaveKeyValue(key, newTrieValue)
		valRecovered, _, err = tdt.RetrieveValue(key)
		assert.Equal(t, newTrieValue, valRecovered)
		assert.Nil(t, err)
	})

	t.Run("nil data trie should err", func(t *testing.T) {
		t.Parallel()

		tdt, err := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
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
		value := append(expectedVal, expectedKey...)
		value = append(value, identifier...)

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, expectedKey) {
					return value, 0, nil
				}
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, enableEpochsHandler)
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, valRecovered)
	})

	t.Run("autoBalance data tries disabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		expectedVal := []byte("value")
		value := append(expectedVal, expectedKey...)
		value = append(value, identifier...)
		hasher := &hashingMocks.HasherMock{}

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, expectedKey) {
					return value, 0, nil
				}
				if bytes.Equal(key, hasher.Compute(string(expectedKey))) {
					assert.Fail(t, "this should not have been called")
				}
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, enableEpochsHandler)
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, valRecovered)
	})

	t.Run("val as struct found in trie", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		expectedVal := []byte("value")
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}

		trie := &trieMock.TrieStub{
			UpdateCalled: func(key, value []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, hasher.Compute(string(expectedKey))) {
					serializedVal, _ := marshaller.Marshal(&dataTrieValue.TrieLeafData{
						Value:   expectedVal,
						Key:     expectedKey,
						Address: identifier,
					})
					return serializedVal, 0, nil
				}
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, hasher, marshaller, enableEpochsHandler)
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, valRecovered)
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
		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

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

		trie := &trieMock.TrieStub{
			UpdateCalled: func(key, value []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(
			identifier,
			hasher,
			marshaller,
			enableEpochsHandler,
		)
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, []byte(nil), valRecovered)
	})

	t.Run("val not found in trie - auto balance disabled", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		hasher := &hashingMocks.HasherMock{}
		marshaller := &marshallerMock.MarshalizerMock{}

		trie := &trieMock.TrieStub{
			UpdateCalled: func(key, value []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
		}
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(
			identifier,
			hasher,
			marshaller,
			enableEpochsHandler,
		)
		assert.NotNil(t, tdt)
		tdt.SetDataTrie(trie)

		valRecovered, _, err := tdt.RetrieveValue(expectedKey)
		assert.Nil(t, err)
		assert.Equal(t, []byte(nil), valRecovered)
	})
}

func TestTrackableDataTrie_SaveDirtyData(t *testing.T) {
	t.Parallel()

	t.Run("no dirty data", func(t *testing.T) {
		t.Parallel()

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

		oldValues, err := tdt.SaveDirtyData(&trieMock.TrieStub{})
		assert.Nil(t, err)
		assert.Equal(t, 0, len(oldValues))
	})

	t.Run("nil trie creates a new trie", func(t *testing.T) {
		t.Parallel()

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
		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

		key := []byte("key")
		_ = tdt.SaveKeyValue(key, []byte("val"))
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, key, oldValues[0].Key)
		assert.Equal(t, []byte(nil), oldValues[0].Value)
		assert.True(t, recreateCalled)
	})

	t.Run("present in trie as valWithAppendedData", func(t *testing.T) {
		t.Parallel()

		identifier := []byte("identifier")
		expectedKey := []byte("key")
		expectedVal := []byte("value")
		value := append(expectedVal, expectedKey...)
		value = append(value, identifier...)
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

		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(key, expectedKey) {
					return value, 0, nil
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
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, hasher, marshaller, enableEpochsHandler)
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, expectedVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, expectedKey, oldValues[0].Key)
		assert.Equal(t, value, oldValues[0].Value)
		assert.True(t, deleteCalled)
		assert.True(t, updateCalled)
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
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, hasher, marshaller, enableEpochsHandler)
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, val)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, expectedKey, oldValues[0].Key)
		assert.Equal(t, expectedVal, oldValues[0].Value)
		assert.True(t, updateCalled)
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
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, hasher, marshaller, enableEpochsHandler)
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, newVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, hasher.Compute(string(expectedKey)), oldValues[0].Key)
		assert.Equal(t, serializedOldTrieVal, oldValues[0].Value)
		assert.True(t, updateCalled)
	})

	t.Run("not present in trie", func(t *testing.T) {
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
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(identifier, hasher, marshaller, enableEpochsHandler)
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, newVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, hasher.Compute(string(expectedKey)), oldValues[0].Key)
		assert.Equal(t, []byte(nil), oldValues[0].Value)
		assert.True(t, updateCalled)
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

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, val)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
	})

	t.Run("nil val autobalance disabled", func(t *testing.T) {
		t.Parallel()

		expectedKey := []byte("key")
		updateCalled := false
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				return []byte("value"), 0, nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, expectedKey, key)
				updateCalled = true
				return nil
			},
		}

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.True(t, updateCalled)
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

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.False(t, deleteCalled)
	})

	t.Run("nil val autobalance enabled, old val saved at hashedKey", func(t *testing.T) {
		t.Parallel()

		hasher := &hashingMocks.HasherMock{}
		expectedKey := []byte("key")
		deleteCalled := false
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(hasher.Compute(string(expectedKey)), key) {
					return []byte("value"), 0, nil
				}

				return nil, 0, nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, hasher.Compute(string(expectedKey)), key)
				deleteCalled = true
				return nil
			},
		}

		enableEpchs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, enableEpchs)
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.True(t, deleteCalled)
	})

	t.Run("nil val autobalance enabled, old val saved at key", func(t *testing.T) {
		t.Parallel()

		expectedKey := []byte("key")
		deleteCalled := 0
		trie := &trieMock.TrieStub{
			GetCalled: func(key []byte) ([]byte, uint32, error) {
				if bytes.Equal(expectedKey, key) {
					return []byte("value"), 0, nil
				}

				return nil, 0, nil
			},
			DeleteCalled: func(key []byte) error {
				assert.Equal(t, expectedKey, key)
				deleteCalled++
				return nil
			},
		}

		enableEpchs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, enableEpchs)
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, nil)
		_, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tdt.DirtyData()))
		assert.Equal(t, 1, deleteCalled)
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
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		tdt, _ := trackableDataTrie.NewTrackableDataTrie(
			identifier,
			hasher,
			marshaller,
			enableEpochsHandler,
		)
		tdt.SetDataTrie(trie)

		_ = tdt.SaveKeyValue(expectedKey, newVal)
		oldValues, err := tdt.SaveDirtyData(trie)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(oldValues))
		assert.Equal(t, expectedKey, oldValues[0].Key)
		assert.Equal(t, []byte(nil), oldValues[0].Value)
		assert.True(t, updateCalled)
	})
}

func TestTrackableDataTrie_MigrateDataTrieLeaves(t *testing.T) {
	t.Parallel()

	t.Run("nil trie", func(t *testing.T) {
		t.Parallel()

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: &trieMock.DataTrieMigratorStub{},
		}
		err := tdt.MigrateDataTrieLeaves(args)
		assert.Equal(t, state.ErrNilTrie, err)
	})

	t.Run("nil trie migrator", func(t *testing.T) {
		t.Parallel()

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		tdt.SetDataTrie(&trieMock.TrieStub{})

		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: nil,
		}
		err := tdt.MigrateDataTrieLeaves(args)
		assert.Equal(t, errorsCommon.ErrNilTrieMigrator, err)
	})

	t.Run("CollectLeavesForMigrationFails", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		tr := &trieMock.TrieStub{
			CollectLeavesForMigrationCalled: func(_ vmcommon.ArgsMigrateDataTrieLeaves) error {
				return expectedErr
			},
		}

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		tdt.SetDataTrie(tr)
		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   core.AutoBalanceEnabled,
			TrieMigrator: &trieMock.DataTrieMigratorStub{},
		}
		err := tdt.MigrateDataTrieLeaves(args)
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
			CollectLeavesForMigrationCalled: func(_ vmcommon.ArgsMigrateDataTrieLeaves) error {
				return nil
			},
		}
		dtm := &trieMock.DataTrieMigratorStub{
			GetLeavesToBeMigratedCalled: func() []core.TrieData {
				return leavesToBeMigrated
			},
		}
		enableEpchs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AutoBalanceDataTriesFlag
			},
		}

		tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, enableEpchs)
		tdt.SetDataTrie(tr)
		args := vmcommon.ArgsMigrateDataTrieLeaves{
			OldVersion:   core.NotSpecified,
			NewVersion:   100,
			TrieMigrator: dtm,
		}
		err := tdt.MigrateDataTrieLeaves(args)
		assert.Nil(t, err)

		dirtyData := tdt.DirtyData()
		assert.Equal(t, len(leavesToBeMigrated), len(dirtyData))
		for i := range leavesToBeMigrated {
			d := dirtyData[string(leavesToBeMigrated[i].Key)]
			assert.Equal(t, leavesToBeMigrated[i].Value, d.Value)
			assert.Equal(t, core.TrieNodeVersion(100), d.NewVersion)
		}
	})
}

func TestTrackableDataTrie_SetAndGetDataTrie(t *testing.T) {
	t.Parallel()

	tdt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("identifier"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

	newTrie := &trieMock.TrieStub{}
	tdt.SetDataTrie(newTrie)
	assert.Equal(t, newTrie, tdt.DataTrie())
}
