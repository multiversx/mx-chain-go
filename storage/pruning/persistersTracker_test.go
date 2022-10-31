package pruning

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
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
	assert.False(t, check.IfNil(pt))
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
