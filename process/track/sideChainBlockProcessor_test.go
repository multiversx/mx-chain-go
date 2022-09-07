package track_test

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSideBlockProcessor_ShouldErrNilBlockProcessor(t *testing.T) {
	t.Parallel()

	scpb, err := track.NewSideChainBlockProcessor(nil)
	assert.Nil(t, scpb)
	assert.Equal(t, process.ErrNilBlockProcessor, err)
}

func TestNewSideBlockProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	scpb, err := track.NewSideChainBlockProcessor(bp)
	assert.NotNil(t, scpb)
	assert.Nil(t, err)
}

func TestSideBlockProcessor_ShouldProcessReceivedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	header := &block.Header{ShardID: 1}
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.SelfNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID != header.GetShardID() {
				return nil, nil, errors.New("wrong shard ID")
			}
			return &block.Header{Nonce: 499}, []byte("hash"), nil
		},
	}
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSideChainBlockProcessor(bp)

	header.Nonce = 499
	assert.False(t, scbp.ShouldProcessReceivedHeader(header))

	header.Nonce = 500
	assert.True(t, scbp.ShouldProcessReceivedHeader(header))
}
