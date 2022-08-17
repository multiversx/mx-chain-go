package pruning

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewTriePersistersTracker(t *testing.T) {
	t.Parallel()

	pt := NewTriePersisterTracker(getArgs())
	assert.NotNil(t, pt)
	assert.Equal(t, int64(7), pt.oldestEpochKeep)
	assert.Equal(t, int64(8), pt.oldestEpochActive)
	assert.Equal(t, 0, pt.numDbsMarkedAsActive)
	assert.Equal(t, 0, pt.numDbsMarkedAsSynced)
}

func TestTriePersistersTracker_HasInitializedEnoughPersisters(t *testing.T) {
	t.Parallel()

	t.Run("test epoch and oldestEpochKeep", func(t *testing.T) {
		t.Parallel()

		pt := NewTriePersisterTracker(getArgs())
		assert.Equal(t, int64(7), pt.oldestEpochKeep)
		pt.numDbsMarkedAsActive = 2

		assert.True(t, pt.HasInitializedEnoughPersisters(6))
		assert.False(t, pt.HasInitializedEnoughPersisters(7))
		assert.False(t, pt.HasInitializedEnoughPersisters(8))
	})

	t.Run("test storageHashBeenSynced", func(t *testing.T) {
		t.Parallel()

		pt := NewTriePersisterTracker(getArgs())
		assert.Equal(t, int64(7), pt.oldestEpochKeep)
		pt.numDbsMarkedAsActive = 1
		pt.numDbsMarkedAsSynced = 1

		assert.True(t, pt.HasInitializedEnoughPersisters(6))
		assert.False(t, pt.HasInitializedEnoughPersisters(7))
		assert.False(t, pt.HasInitializedEnoughPersisters(8))

		pt.numDbsMarkedAsActive = 0
		assert.False(t, pt.HasInitializedEnoughPersisters(6))

		pt.numDbsMarkedAsActive = 1
		pt.numDbsMarkedAsSynced = 0
		assert.False(t, pt.HasInitializedEnoughPersisters(6))
	})

	t.Run("test hasActiveDbsNecessary", func(t *testing.T) {
		t.Parallel()

		pt := NewTriePersisterTracker(getArgs())
		assert.Equal(t, int64(7), pt.oldestEpochKeep)

		assert.False(t, pt.HasInitializedEnoughPersisters(6))
		assert.False(t, pt.HasInitializedEnoughPersisters(7))
		assert.False(t, pt.HasInitializedEnoughPersisters(8))

		pt.numDbsMarkedAsActive = 1
		assert.False(t, pt.HasInitializedEnoughPersisters(6))

		pt.numDbsMarkedAsActive = 2
		assert.True(t, pt.HasInitializedEnoughPersisters(6))

		pt.numDbsMarkedAsActive = 3
		assert.True(t, pt.HasInitializedEnoughPersisters(6))
	})
}

func TestTriePersistersTracker_ShouldClosePersister(t *testing.T) {
	t.Parallel()

	t.Run("increases numDbsMarkedAsActive", func(t *testing.T) {
		t.Parallel()

		p := &mock.PersisterStub{
			GetCalled: func(key []byte) ([]byte, error) {
				if bytes.Equal(key, []byte(common.ActiveDBKey)) {
					return []byte(common.ActiveDBVal), nil
				}
				return nil, nil
			},
		}
		pt := NewTriePersisterTracker(getArgs())
		assert.Equal(t, int64(8), pt.oldestEpochActive)
		assert.Equal(t, 0, pt.numDbsMarkedAsSynced)
		assert.Equal(t, 0, pt.numDbsMarkedAsActive)

		_ = pt.ShouldClosePersister(p, 0)
		assert.Equal(t, 0, pt.numDbsMarkedAsSynced)
		assert.Equal(t, 1, pt.numDbsMarkedAsActive)
	})

	t.Run("increases numDbsMarkedAsSynced", func(t *testing.T) {
		t.Parallel()

		p := &mock.PersisterStub{
			GetCalled: func(key []byte) ([]byte, error) {
				if bytes.Equal(key, []byte(common.TrieSyncedKey)) {
					return []byte(common.TrieSyncedVal), nil
				}
				return nil, nil
			},
		}
		pt := NewTriePersisterTracker(getArgs())
		assert.Equal(t, int64(8), pt.oldestEpochActive)
		assert.Equal(t, 0, pt.numDbsMarkedAsSynced)
		assert.Equal(t, 0, pt.numDbsMarkedAsActive)

		_ = pt.ShouldClosePersister(p, 0)
		assert.Equal(t, 1, pt.numDbsMarkedAsSynced)
		assert.Equal(t, 0, pt.numDbsMarkedAsActive)
	})

	t.Run("should close persister", func(t *testing.T) {
		t.Parallel()

		p := &mock.PersisterStub{}
		pt := NewTriePersisterTracker(getArgs())
		assert.Equal(t, int64(8), pt.oldestEpochActive)

		assert.Equal(t, 0, pt.numDbsMarkedAsActive)
		assert.False(t, pt.ShouldClosePersister(p, 7))

		pt.numDbsMarkedAsActive = 2
		assert.True(t, pt.ShouldClosePersister(p, 7))

		assert.False(t, pt.ShouldClosePersister(p, 8))
	})
}
