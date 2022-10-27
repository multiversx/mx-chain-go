package trie_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
)

func getTrieStorageManagerOptions() trie.StorageManagerOptions {
	return trie.StorageManagerOptions{
		PruningEnabled:     true,
		SnapshotsEnabled:   true,
		CheckpointsEnabled: true,
	}
}

func TestTrieFactory_CreateWithoutPruning(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.PruningEnabled = false
	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutPruning", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateWithoutSnapshot(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.SnapshotsEnabled = false
	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutSnapshot", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateWithoutCheckpoints(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.CheckpointsEnabled = false
	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutCheckpoints", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateNormal(t *testing.T) {
	t.Parallel()

	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), getTrieStorageManagerOptions())
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
	tsm = &testscommon.StorageManagerStub{
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
		TakeSnapshotCalled: func(_ []byte, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
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
		AddDirtyCheckpointHashesCalled: func(_ []byte, _ common.ModifiedHashes) bool {
			return true
		},
		GetBaseTrieStorageManagerCalled: func() common.StorageManager {
			tsm, _ = trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
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

	// NewTrieStorageManagerWithoutCheckpoints testing
	tsm, err = trie.NewTrieStorageManagerWithoutCheckpoints(tsm)
	assert.Nil(t, err)

	testTsmWithoutPruning(t, tsm)

	getCalled = false
	testTsmWithoutSnapshot(t, tsm, returnedVal)
	assert.Equal(t, 4, putCalled)
	assert.True(t, getCalled)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	tsm.SetCheckpoint(nil, nil, iteratorChannels, nil, &trieMock.MockStatistics{})

	select {
	case <-iteratorChannels.LeavesChan:
	default:
		assert.Fail(t, "unclosed channel")
	}

	assert.False(t, tsm.AddDirtyCheckpointHashes([]byte("hash"), make(map[string]struct{})))
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
		ErrChan:    make(chan error, 1),
	}
	tsm.TakeSnapshot(nil, nil, nil, iteratorChannels, nil, &trieMock.MockStatistics{}, 10)

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
