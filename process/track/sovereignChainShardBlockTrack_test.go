package track_test

import (
	"errors"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSovereignChainShardBlockTrack_ShouldErrNilBlockTracker(t *testing.T) {
	t.Parallel()

	scsbt, err := track.NewSovereignChainShardBlockTrack(nil)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewSovereignChainShardBlockTrack_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	sbt.SetBlockProcessor(nil)

	scsbt, err := track.NewSovereignChainShardBlockTrack(sbt)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestNewSovereignChainShardBlockTrack_ShouldWork(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	scsbt, err := track.NewSovereignChainShardBlockTrack(sbt)
	assert.NotNil(t, scsbt)
	assert.Nil(t, err)
}

func TestSovereignChainShardBlockTrack_ComputeLongestSelfChainShouldWork(t *testing.T) {
	t.Parallel()

	lastNotarizedHeader := &block.Header{Nonce: 1}
	lastNotarizedHash := []byte("hash")
	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinatorMock.CurrentShard = 1
	shardBlockTrackArguments.ShardCoordinator = shardCoordinatorMock
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)
	selfNotarizer := &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID != shardCoordinatorMock.CurrentShard {
				return nil, nil, errors.New("wrong shard ID")
			}
			return lastNotarizedHeader, lastNotarizedHash, nil
		},
	}
	sbt.SetSelfNotarizer(selfNotarizer)
	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header, hash, headers, hashes := scsbt.ComputeLongestSelfChain()

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Equal(t, 0, len(headers))
	assert.Equal(t, 0, len(hashes))
}

func TestSovereignChainShardBlockTrack_GetSelfNotarizedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	lastNotarizedHeader := &block.Header{Nonce: 1}
	lastNotarizedHash := []byte("hash")
	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinatorMock.CurrentShard = 1
	shardBlockTrackArguments.ShardCoordinator = shardCoordinatorMock
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)
	selfNotarizer := &mock.BlockNotarizerHandlerMock{
		GetNotarizedHeaderCalled: func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
			if shardID != shardCoordinatorMock.CurrentShard {
				return nil, nil, errors.New("wrong shard ID")
			}
			return lastNotarizedHeader, lastNotarizedHash, nil
		},
	}
	sbt.SetSelfNotarizer(selfNotarizer)
	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header, hash, err := scsbt.GetSelfNotarizedHeader(core.MetachainShardId, 0)

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Nil(t, err)
}
