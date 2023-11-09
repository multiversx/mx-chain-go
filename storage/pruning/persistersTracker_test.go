package pruning

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getArgs() EpochArgs {
	return EpochArgs{
		StartingEpoch:         10,
		NumOfEpochsToKeep:     4,
		NumOfActivePersisters: 3,
	}
}

func TestNewPersistersTracker(t *testing.T) {
	t.Parallel()

	pt := NewPersistersTracker(getArgs())
	assert.NotNil(t, pt)
	assert.Equal(t, int64(7), pt.oldestEpochKeep)
	assert.Equal(t, int64(8), pt.oldestEpochActive)
}

func TestPersistersTracker_HasInitializedEnoughPersisters(t *testing.T) {
	t.Parallel()

	pt := NewPersistersTracker(getArgs())

	assert.False(t, pt.HasInitializedEnoughPersisters(7))
	assert.True(t, pt.HasInitializedEnoughPersisters(6))
}

func TestPersistersTracker_ShouldClosePersister(t *testing.T) {
	t.Parallel()

	pt := NewPersistersTracker(getArgs())
	assert.NotNil(t, pt)
	assert.Equal(t, int64(8), pt.oldestEpochActive)

	assert.False(t, pt.ShouldClosePersister(8))
	assert.True(t, pt.ShouldClosePersister(7))
}

func TestPersistersTracker_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var npt *normalPersistersTracker
	require.True(t, npt.IsInterfaceNil())

	npt = NewPersistersTracker(getArgs())
	require.False(t, npt.IsInterfaceNil())
}
