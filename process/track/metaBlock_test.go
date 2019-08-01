package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/stretchr/testify/assert"
)

func TestMetaBlockTracker_NewMetaBlockTrackerShouldWork(t *testing.T) {
	t.Parallel()

	mbt, err := track.NewMetaBlockTracker()
	assert.Nil(t, err)
	assert.NotNil(t, mbt)
}

func TestMetaBlockTracker_UnnotarisedBlocksShouldWork(t *testing.T) {
	t.Parallel()

	mbt, _ := track.NewMetaBlockTracker()
	headers := mbt.UnnotarisedBlocks()
	assert.Equal(t, make([]data.HeaderHandler, 0), headers)
}

func TestMetaBlockTracker_BlockBroadcastRoundShouldWork(t *testing.T) {
	t.Parallel()

	mbt, _ := track.NewMetaBlockTracker()
	assert.Equal(t, int64(0), mbt.BlockBroadcastRound(1))
}

func TestMetaBlockTracker_RemoveNotarisedBlocksShouldWork(t *testing.T) {
	t.Parallel()

	mbt, _ := track.NewMetaBlockTracker()
	err := mbt.RemoveNotarisedBlocks(&block.MetaBlock{})
	assert.Nil(t, err)
}
