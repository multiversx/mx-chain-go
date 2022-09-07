package track_test

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSideChainShardBlockTrack_ShouldErrNilBlockTracker(t *testing.T) {
	t.Parallel()

	scsbt, err := track.NewSideChainShardBlockTrack(nil)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewSideChainShardBlockTrack_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	sbt.SetBlockProcessor(nil)

	scsbt, err := track.NewSideChainShardBlockTrack(sbt)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestNewSideChainShardBlockTrack_ShouldWork(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	scsbt, err := track.NewSideChainShardBlockTrack(sbt)
	assert.NotNil(t, scsbt)
	assert.Nil(t, err)
}

func TestSideChainShardBlockTrack_ComputeLongestSelfChainShouldWork(t *testing.T) {
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
	scsbt, _ := track.NewSideChainShardBlockTrack(sbt)

	header, hash, headers, hashes := scsbt.ComputeLongestSelfChainFunc()

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Equal(t, 0, len(headers))
	assert.Equal(t, 0, len(hashes))
}

func TestSideChainShardBlockTrack_GetSelfNotarizedHeaderShouldWork(t *testing.T) {
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
	scsbt, _ := track.NewSideChainShardBlockTrack(sbt)

	header, hash, err := scsbt.GetSelfNotarizedHeaderFunc(core.MetachainShardId, 0)

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Nil(t, err)
}
