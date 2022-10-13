package sync

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewSideChainShardForkDetector_ShouldErrNilForkDetector(t *testing.T) {
	t.Parallel()

	scsfd, err := NewSideChainShardForkDetector(nil)
	assert.Nil(t, scsfd)
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

	scsfd, err := NewSideChainShardForkDetector(sfd)
	assert.NotNil(t, scsfd)
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

	scsfd, _ := NewSideChainShardForkDetector(sfd)

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
