package sync

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSideChainShardForkDetector_ShouldErrNilForkDetector(t *testing.T) {
	t.Parallel()

	sdsfd, err := NewSideChainShardForkDetector(nil)
	assert.Nil(t, sdsfd)
	assert.Equal(t, process.ErrNilForkDetector, err)
}

func TestNewSideChainShardForkDetector_ShouldWork(t *testing.T) {
	t.Parallel()

	sfd, _ := NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	sdsfd, err := NewSideChainShardForkDetector(sfd)
	assert.NotNil(t, sdsfd)
	assert.Nil(t, err)
}

func TestSideChainShardForkDetector_DoJobOnBHProcessedShouldWork(t *testing.T) {
	t.Parallel()

	sfd, _ := NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	sdsfd, _ := NewSideChainShardForkDetector(sfd)

	finalCheckpoint := &checkpointInfo{
		nonce: 1,
		round: 1,
		hash:  []byte("hash1"),
	}
	sdsfd.addCheckpoint(finalCheckpoint)

	header := &block.Header{Round: 2, Nonce: 2}
	headerHash := []byte("hash2")
	sdsfd.doJobOnBHProcessed(header, headerHash, nil, nil)

	expectedCheckpoint := &checkpointInfo{
		nonce: header.Nonce,
		round: header.Round,
		hash:  headerHash,
	}

	assert.Equal(t, finalCheckpoint, sdsfd.finalCheckpoint())
	assert.Equal(t, expectedCheckpoint, sdsfd.lastCheckpoint())
}
