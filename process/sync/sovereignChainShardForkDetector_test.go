package sync

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignChainShardForkDetector_ShouldErrNilForkDetector(t *testing.T) {
	t.Parallel()

	scsfd, err := NewSovereignChainShardForkDetector(nil)
	assert.Nil(t, scsfd)
	assert.Equal(t, process.ErrNilForkDetector, err)
}

func TestNewSovereignChainShardForkDetector_ShouldWork(t *testing.T) {
	t.Parallel()

	sfd, _ := NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	scsfd, err := NewSovereignChainShardForkDetector(sfd)
	assert.NotNil(t, scsfd)
	assert.Nil(t, err)
}

func TestSovereignChainShardForkDetector_DoJobOnBHProcessedShouldWork(t *testing.T) {
	t.Parallel()

	sfd, _ := NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	scsfd, _ := NewSovereignChainShardForkDetector(sfd)

	finalCheckpoint := &checkpointInfo{
		nonce: 1,
		round: 1,
		hash:  []byte("hash1"),
	}
	scsfd.addCheckpoint(finalCheckpoint)

	header := &block.Header{Round: 2, Nonce: 2}
	headerHash := []byte("hash2")
	scsfd.doJobOnBHProcessedFunc(header, headerHash, nil, nil)

	expectedCheckpoint := &checkpointInfo{
		nonce: header.Nonce,
		round: header.Round,
		hash:  headerHash,
	}

	assert.Equal(t, finalCheckpoint, scsfd.finalCheckpoint())
	assert.Equal(t, expectedCheckpoint, scsfd.lastCheckpoint())
}
