package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockBalancer_ShouldWork(t *testing.T) {
	t.Parallel()

	bb, err := track.NewBlockBalancer()

	assert.Nil(t, err)
	assert.NotNil(t, bb)
}

func TestSetAndGetNumPendingMiniBlocks_ShouldWork(t *testing.T) {
	t.Parallel()

	bb, _ := track.NewBlockBalancer()

	bb.SetNumPendingMiniBlocks(0, 2)
	bb.SetNumPendingMiniBlocks(1, 3)

	assert.Equal(t, uint32(2), bb.GetNumPendingMiniBlocks(0))
	assert.Equal(t, uint32(3), bb.GetNumPendingMiniBlocks(1))
	assert.Equal(t, uint32(0), bb.GetNumPendingMiniBlocks(2))
}

func TestSetAndGetLastShardProcessedMetaNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	bb, _ := track.NewBlockBalancer()

	bb.SetLastShardProcessedMetaNonce(0, 2)
	bb.SetLastShardProcessedMetaNonce(1, 3)

	assert.Equal(t, uint64(2), bb.GetLastShardProcessedMetaNonce(0))
	assert.Equal(t, uint64(3), bb.GetLastShardProcessedMetaNonce(1))
	assert.Equal(t, uint64(0), bb.GetLastShardProcessedMetaNonce(2))
}
