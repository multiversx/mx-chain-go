package trie_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

func getTrieStorageManagerOptions() trie.StorageManagerOptions {
	return trie.StorageManagerOptions{
		PruningEnabled:   true,
		SnapshotsEnabled: true,
	}
}

func TestTrieFactory_CreateWithoutPruning(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.PruningEnabled = false
	tsm, err := trie.CreateTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutPruning", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateWithoutSnapshot(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.SnapshotsEnabled = false
	tsm, err := trie.CreateTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutSnapshot", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateNormal(t *testing.T) {
	t.Parallel()

	tsm, err := trie.CreateTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters(), getTrieStorageManagerOptions())
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManager", fmt.Sprintf("%T", tsm))
}

func TestTrieStorageManager_SerialFuncShadowingCallsExpectedImpl(t *testing.T) {
	t.Parallel()

	var tsm common.StorageManager
	shouldNotHaveBeenCalledErr := fmt.Errorf("remove should not have been called")
	getCalled := false
	returnedVal := []byte("existingVal")
	putCalled := 0
	tsm = &storageManager.StorageManagerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			getCalled = true
			return returnedVal, nil
		},
		PutCalled: func(_ []byte, _ []byte) error {
			putCalled++
			return nil
		},
		RemoveCalled: func(_ []byte) error {
			return shouldNotHaveBeenCalledErr
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
		TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
			assert.Fail(t, shouldNotHaveBeenCalledErr.Error())
		},
		GetLatestStorageEpochCalled: func() (uint32, error) {
			return 6, nil
		},
		SetEpochForPutOperationCalled: func(_ uint32) {
			assert.Fail(t, shouldNotHaveBeenCalledErr.Error())
		},
		ShouldTakeSnapshotCalled: func() bool {
			return true
		},
		GetBaseTrieStorageManagerCalled: func() common.StorageManager {
			tsm, _ = trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
			return tsm
		},
	}

	err := tsm.Remove([]byte("hash"))
	assert.Equal(t, shouldNotHaveBeenCalledErr, err)

	// NewTrieStorageManagerWithoutPruning testing
	tsm, err = trie.NewTrieStorageManagerWithoutPruning(tsm)
	assert.Nil(t, err)

	testTsmWithoutPruning(t, tsm)

	// NewTrieStorageManagerWithoutSnapshot testing
	tsm, err = trie.NewTrieStorageManagerWithoutSnapshot(tsm)
	assert.Nil(t, err)

	testTsmWithoutPruning(t, tsm)

	testTsmWithoutSnapshot(t, tsm, returnedVal)
	assert.Equal(t, 2, putCalled)
	assert.True(t, getCalled)

	getCalled = false
	testTsmWithoutSnapshot(t, tsm, returnedVal)
	assert.Equal(t, 4, putCalled)
	assert.True(t, getCalled)
}

func testTsmWithoutPruning(t *testing.T, tsm common.StorageManager) {
	err := tsm.Remove([]byte("hash"))
	assert.Nil(t, err)
	assert.False(t, tsm.IsPruningEnabled())
}

func testTsmWithoutSnapshot(
	t *testing.T,
	tsm common.StorageManager,
	returnedVal []byte,
) {
	val, err := tsm.GetFromCurrentEpoch([]byte("hash"))
	assert.Nil(t, err)
	assert.Equal(t, returnedVal, val)

	_ = tsm.PutInEpoch([]byte("hash"), []byte("val"), 0)
	_ = tsm.PutInEpochWithoutCache([]byte("hash"), []byte("val"), 0)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	tsm.TakeSnapshot("", nil, nil, iteratorChannels, nil, &trieMock.MockStatistics{}, 10)

	select {
	case <-iteratorChannels.LeavesChan:
	default:
		assert.Fail(t, "unclosed channel")
	}

	epoch, err := tsm.GetLatestStorageEpoch()
	assert.Equal(t, uint32(0), epoch)
	assert.Nil(t, err)

	tsm.SetEpochForPutOperation(5)
	assert.False(t, tsm.ShouldTakeSnapshot())
}
