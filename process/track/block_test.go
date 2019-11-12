package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockTrack_ShouldErrNilRounder(t *testing.T) {
	bt, err := track.NewBlockTrack(nil)

	assert.Equal(t, process.ErrNilRounder, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldWork(t *testing.T) {
	bt, err := track.NewBlockTrack(&mock.RounderMock{})

	assert.Nil(t, err)
	assert.NotNil(t, bt)
}

func TestNewBlockTrack_SetLastHeaderForShardShouldNotUpdateIfHeaderIsNil(t *testing.T) {
	bt, _ := track.NewBlockTrack(&mock.RounderMock{})
	bt.SetLastHeaderForShard(nil)
	lastHeader := bt.LastHeaderForShard(0)
	assert.Nil(t, lastHeader)
}

func TestNewBlockTrack_SetLastHeaderForShardShouldNotUpdateIfRoundIsLowerOrEqual(t *testing.T) {
	bt, _ := track.NewBlockTrack(&mock.RounderMock{})

	header1 := &block.Header{Round: 2}
	header2 := &block.Header{Round: 1}
	header3 := &block.Header{Round: 2}

	bt.SetLastHeaderForShard(header1)
	lastHeader := bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)

	bt.SetLastHeaderForShard(header2)
	lastHeader = bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)

	bt.SetLastHeaderForShard(header3)
	lastHeader = bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)
}

func TestNewBlockTrack_SetLastHeaderForShardShouldUpdate(t *testing.T) {
	bt, _ := track.NewBlockTrack(&mock.RounderMock{})

	header1 := &block.Header{Round: 2}
	header2 := &block.Header{Round: 3}
	header3 := &block.Header{Round: 4}

	header4 := &block.MetaBlock{Round: 2}
	header5 := &block.MetaBlock{Round: 3}
	header6 := &block.MetaBlock{Round: 4}

	bt.SetLastHeaderForShard(header1)
	bt.SetLastHeaderForShard(header4)
	lastHeader := bt.LastHeaderForShard(0)
	lastHeaderMeta := bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header1, lastHeader)
	assert.Equal(t, header4, lastHeaderMeta)

	bt.SetLastHeaderForShard(header2)
	bt.SetLastHeaderForShard(header5)
	lastHeader = bt.LastHeaderForShard(0)
	lastHeaderMeta = bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header2, lastHeader)
	assert.Equal(t, header5, lastHeaderMeta)

	bt.SetLastHeaderForShard(header3)
	bt.SetLastHeaderForShard(header6)
	lastHeader = bt.LastHeaderForShard(0)
	lastHeaderMeta = bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header3, lastHeader)
	assert.Equal(t, header6, lastHeaderMeta)
}

func TestNewBlockTrack_LastHeaderForShardShouldWork(t *testing.T) {
	bt, _ := track.NewBlockTrack(&mock.RounderMock{})

	header1 := &block.Header{ShardId: 0, Round: 2}
	header2 := &block.Header{ShardId: 1, Round: 2}
	header3 := &block.MetaBlock{Round: 2}

	bt.SetLastHeaderForShard(header1)
	bt.SetLastHeaderForShard(header2)
	bt.SetLastHeaderForShard(header3)

	lastHeader := bt.LastHeaderForShard(header1.GetShardID())
	assert.Equal(t, header1, lastHeader)

	lastHeader = bt.LastHeaderForShard(header2.GetShardID())
	assert.Equal(t, header2, lastHeader)

	lastHeader = bt.LastHeaderForShard(header3.GetShardID())
	assert.Equal(t, header3, lastHeader)
}

func TestNewBlockTrack_IsShardStuckShoudReturnFalseWhenListIsEmpty(t *testing.T) {
	bt, _ := track.NewBlockTrack(&mock.RounderMock{})

	isShardStuck := bt.IsShardStuck(0)
	assert.False(t, isShardStuck)
}

func TestNewBlockTrack_IsShardStuckShoudReturnFalse(t *testing.T) {
	bt, _ := track.NewBlockTrack(&mock.RounderMock{})

	bt.SetLastHeaderForShard(&block.Header{})
	isShardStuck := bt.IsShardStuck(0)
	assert.False(t, isShardStuck)
}

func TestNewBlockTrack_IsShardStuckShoudReturnTrue(t *testing.T) {
	rounderMock := &mock.RounderMock{}
	bt, _ := track.NewBlockTrack(rounderMock)

	bt.SetLastHeaderForShard(&block.Header{})
	rounderMock.RoundIndex = process.MaxRoundsWithoutCommittedBlock + 1
	isShardStuck := bt.IsShardStuck(0)
	assert.True(t, isShardStuck)
}
