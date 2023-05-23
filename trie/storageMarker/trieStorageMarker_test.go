package storageMarker

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestTrieStorageMarker_MarkStorerAsSyncedAndActive(t *testing.T) {
	t.Parallel()

	t.Run("all operations error should work", func(t *testing.T) {
		t.Parallel()

		sm := NewTrieStorageMarker()
		assert.NotNil(t, sm)

		expectedErr := errors.New("expected err")

		getLatestStorageEpochCalled := false
		putCalled := false
		putInEpochWithoutCacheCalled := false
		storer := &testscommon.StorageManagerStub{
			GetLatestStorageEpochCalled: func() (uint32, error) {
				getLatestStorageEpochCalled = true
				return 0, expectedErr
			},
			PutCalled: func(key []byte, val []byte) error {
				putCalled = true
				return expectedErr
			},
			PutInEpochWithoutCacheCalled: func(key []byte, val []byte, epoch uint32) error {
				putInEpochWithoutCacheCalled = true
				return expectedErr
			},
		}
		sm.MarkStorerAsSyncedAndActive(storer)
		assert.True(t, getLatestStorageEpochCalled)
		assert.True(t, putCalled)
		assert.True(t, putInEpochWithoutCacheCalled)
	})
	t.Run("mark storer as synced and active epoch 5", func(t *testing.T) {
		sm := NewTrieStorageMarker()
		assert.NotNil(t, sm)

		trieSyncedKeyPut := false
		activeDbKeyPut := false
		storer := &testscommon.StorageManagerStub{
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return 5, nil
			},
			PutCalled: func(key []byte, val []byte) error {
				assert.Equal(t, []byte(common.TrieSyncedKey), key)
				assert.Equal(t, []byte(common.TrieSyncedVal), val)
				trieSyncedKeyPut = true
				return nil
			},
			PutInEpochWithoutCacheCalled: func(key []byte, val []byte, epoch uint32) error {
				assert.Equal(t, []byte(common.ActiveDBKey), key)
				assert.Equal(t, []byte(common.ActiveDBVal), val)
				assert.Equal(t, uint32(4), epoch)
				activeDbKeyPut = true
				return nil
			},
		}
		sm.MarkStorerAsSyncedAndActive(storer)
		assert.True(t, trieSyncedKeyPut)
		assert.True(t, activeDbKeyPut)
	})
	t.Run("mark storer as synced and active epoch 0", func(t *testing.T) {
		sm := NewTrieStorageMarker()
		assert.NotNil(t, sm)

		trieSyncedKeyPut := false
		activeDbKeyPut := false
		storer := &testscommon.StorageManagerStub{
			GetLatestStorageEpochCalled: func() (uint32, error) {
				return 0, nil
			},
			PutCalled: func(key []byte, val []byte) error {
				assert.Equal(t, []byte(common.TrieSyncedKey), key)
				assert.Equal(t, []byte(common.TrieSyncedVal), val)
				trieSyncedKeyPut = true
				return nil
			},
			PutInEpochWithoutCacheCalled: func(key []byte, val []byte, epoch uint32) error {
				assert.Equal(t, []byte(common.ActiveDBKey), key)
				assert.Equal(t, []byte(common.ActiveDBVal), val)
				assert.Equal(t, uint32(0), epoch)
				activeDbKeyPut = true
				return nil
			},
		}
		sm.MarkStorerAsSyncedAndActive(storer)
		assert.True(t, trieSyncedKeyPut)
		assert.True(t, activeDbKeyPut)
	})
}
