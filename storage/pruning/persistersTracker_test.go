package pruning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getArgs() *EpochArgs {
	return &EpochArgs{
		StartingEpoch:         10,
		NumOfEpochsToKeep:     4,
		NumOfActivePersisters: 3,
	}
}

func TestNewPersistersTracker(t *testing.T) {
	t.Parallel()

	pt := newPersistersTracker(getArgs())
	assert.NotNil(t, pt)
	assert.Equal(t, int64(7), pt.oldestEpochKeep)
	assert.Equal(t, int64(8), pt.oldestEpochActive)
}

func TestPersistersTracker_hasInitializedEnoughPersisters(t *testing.T) {
	t.Parallel()

	pt := newPersistersTracker(getArgs())

	assert.False(t, pt.hasInitializedEnoughPersisters(7))
	assert.True(t, pt.hasInitializedEnoughPersisters(6))
}

func TestPersistersTracker_shouldClosePersister(t *testing.T) {
	t.Parallel()

	pt := newPersistersTracker(getArgs())
	assert.NotNil(t, pt)
	assert.Equal(t, int64(8), pt.oldestEpochActive)

	assert.False(t, pt.shouldClosePersister(8))
	assert.True(t, pt.shouldClosePersister(7))
}
