package lastSnapshotMarker

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/stretchr/testify/assert"
)

func TestNewLastSnapshotMarker(t *testing.T) {
	t.Parallel()

	var lsm *lastSnapshotMarker
	assert.True(t, lsm.IsInterfaceNil())

	lsm = NewLastSnapshotMarker()
	assert.False(t, lsm.IsInterfaceNil())
}

func TestLastSnapshotMarker_AddMarker(t *testing.T) {
	t.Parallel()

	t.Run("err waiting for storage epoch change", func(t *testing.T) {
		t.Parallel()

		trieStorageManager := &storageManager.StorageManagerStub{
			IsClosedCalled: func() bool {
				return true
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "should not have been called")
				return nil
			},
		}

		lsm := NewLastSnapshotMarker()
		lsm.AddMarker(trieStorageManager, 1, []byte("rootHash"))
	})
	t.Run("epoch <= latestFinishedSnapshotEpoch", func(t *testing.T) {
		t.Parallel()

		trieStorageManager := &storageManager.StorageManagerStub{
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "should not have been called")
				return nil
			},
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return 1, nil
			},
		}

		lsm := NewLastSnapshotMarker()
		lsm.latestFinishedSnapshotEpoch = 2
		lsm.AddMarker(trieStorageManager, 1, []byte("rootHash"))
	})
	t.Run("lastSnapshot is saved in epoch", func(t *testing.T) {
		t.Parallel()

		val := []byte("rootHash")
		epoch := uint32(1)
		putInEpochCalled := false
		trieStorageManager := &storageManager.StorageManagerStub{
			PutInEpochCalled: func(key []byte, v []byte, e uint32) error {
				putInEpochCalled = true
				assert.Equal(t, []byte(lastSnapshot), key)
				assert.Equal(t, val, v)
				assert.Equal(t, epoch, e)
				return nil
			},
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return epoch, nil
			},
		}

		lsm := NewLastSnapshotMarker()
		lsm.AddMarker(trieStorageManager, epoch, val)
		assert.True(t, putInEpochCalled)
	})
}

func TestLastSnapshotMarker_RemoveMarker(t *testing.T) {
	t.Parallel()

	removeIsCalled := false
	trieStorageManager := &storageManager.StorageManagerStub{
		RemoveFromAllActiveEpochsCalled: func(_ []byte) error {
			removeIsCalled = true
			return nil
		},
	}

	lsm := NewLastSnapshotMarker()
	lsm.RemoveMarker(trieStorageManager, 5, []byte("rootHash"))
	assert.True(t, removeIsCalled)
	assert.Equal(t, uint32(5), lsm.latestFinishedSnapshotEpoch)
}

func TestLastSnapshotMarker_GetMarkerInfo(t *testing.T) {
	t.Parallel()

	getCalled := false
	rootHash := []byte("rootHash")
	trieStorageManager := &storageManager.StorageManagerStub{
		GetFromCurrentEpochCalled: func(bytes []byte) ([]byte, error) {
			getCalled = true
			assert.Equal(t, []byte(lastSnapshot), bytes)
			return rootHash, nil
		},
	}

	lsm := NewLastSnapshotMarker()
	val, err := lsm.GetMarkerInfo(trieStorageManager)
	assert.Nil(t, err)
	assert.True(t, getCalled)
	assert.Equal(t, rootHash, val)
}
