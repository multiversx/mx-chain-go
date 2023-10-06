package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/stretchr/testify/assert"
)

func getDefaultArgs() StorageEpochChangeWaitArgs {
	return StorageEpochChangeWaitArgs{
		Epoch:                         1,
		WaitTimeForSnapshotEpochCheck: time.Millisecond * 100,
		SnapshotWaitTimeout:           time.Second,
		TrieStorageManager:            &storageManager.StorageManagerStub{},
	}
}

func TestSnapshotsManager_WaitForStorageEpochChange(t *testing.T) {
	t.Parallel()

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.SnapshotWaitTimeout = time.Millisecond

		err := WaitForStorageEpochChange(args)
		assert.Error(t, err)
	})
	t.Run("getLatestStorageEpoch error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("getLatestStorageEpoch error")

		args := getDefaultArgs()
		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return 0, expectedError
			},
		}

		err := WaitForStorageEpochChange(args)
		assert.Equal(t, expectedError, err)
	})
	t.Run("storage manager closed error", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return 0, nil
			},
			IsClosedCalled: func() bool {
				return true
			},
		}

		err := WaitForStorageEpochChange(args)
		assert.Equal(t, core.ErrContextClosing, err)
	})
	t.Run("storage epoch change timeout", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.WaitTimeForSnapshotEpochCheck = time.Millisecond
		args.SnapshotWaitTimeout = time.Millisecond * 5
		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return 0, nil
			},
		}

		err := WaitForStorageEpochChange(args)
		assert.Error(t, err)
	})
	t.Run("returns when latestStorageEpoch == snapshotEpoch", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return 1, nil
			},
		}

		err := WaitForStorageEpochChange(args)
		assert.Nil(t, err)
	})
}
