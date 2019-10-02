package sync_test

import (
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewBasicForkDetector_ShouldErrNilRounder(t *testing.T) {
	t.Parallel()
	bfd, err := sync.NewShardForkDetector(nil)
	assert.Equal(t, process.ErrNilRounder, err)
	assert.Nil(t, bfd)
}

func TestNewBasicForkDetector_ShouldWork(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, err := sync.NewShardForkDetector(rounderMock)
	assert.Nil(t, err)
	assert.NotNil(t, bfd)
}

func TestBasicForkDetector_AddHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.AddHeader(nil, make([]byte, 0), process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHeader, err)
}

func TestBasicForkDetector_AddHeaderNilHashShouldErr(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.AddHeader(&block.Header{}, nil, process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHash, err)
}

func TestBasicForkDetector_AddHeaderUnsignedBlockShouldErr(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 1},
		make([]byte, 0),
		process.BHProcessed,
		nil,
		nil)
	assert.Equal(t, sync.ErrBlockIsNotSigned, err)
}

func TestBasicForkDetector_AddHeaderNotPresentShouldWork(t *testing.T) {
	t.Parallel()
	hdr := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := make([]byte, 0)
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)

	err := bfd.AddHeader(hdr, hash, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, hash, hInfos[0].Hash())
}

func TestBasicForkDetector_AddHeaderPresentShouldAppend(t *testing.T) {
	t.Parallel()
	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)

	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	err := bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 2, len(hInfos))
	assert.Equal(t, hash1, hInfos[0].Hash())
	assert.Equal(t, hash2, hInfos[1].Hash())
}

func TestBasicForkDetector_AddHeaderWithProcessedBlockShouldSetCheckpoint(t *testing.T) {
	t.Parallel()
	hdr1 := &block.Header{Nonce: 69, Round: 72, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	rounderMock := &mock.RounderMock{RoundIndex: 73}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, hdr1.Nonce, bfd.LastCheckpointNonce())
}

func TestBasicForkDetector_AddHeaderPresentShouldRewriteState(t *testing.T) {
	t.Parallel()
	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)

	_ = bfd.AddHeader(hdr1, hash, process.BHReceived, nil, nil)
	err := bfd.AddHeader(hdr2, hash, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, hash, hInfos[0].Hash())
	assert.Equal(t, process.BHProcessed, hInfos[0].GetBlockHeaderState())
}

func TestBasicForkDetector_CheckBlockValidityShouldErrLowerRoundInBlock(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	bfd.SetFinalCheckpoint(1, 1)
	err := bfd.CheckBlockValidity(&block.Header{PubKeysBitmap: []byte("X")}, process.BHProcessed)
	assert.Equal(t, sync.ErrLowerRoundInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrLowerNonceInBlock(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	bfd.SetFinalCheckpoint(1, 1)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")}, process.BHProcessed)
	assert.Equal(t, sync.ErrLowerNonceInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrHigherRoundInBlock(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 0}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")}, process.BHProcessed)
	assert.Equal(t, sync.ErrHigherRoundInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrHigherNonceInBlock(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 2, Round: 1, PubKeysBitmap: []byte("X")}, process.BHProcessed)
	assert.Equal(t, sync.ErrHigherNonceInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrRandomSeedIsNotValid(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 1}, process.BHProposed)
	assert.Equal(t, sync.ErrRandomSeedNotValid, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrBlockIsNotSigned(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 1}, process.BHProcessed)
	assert.Equal(t, sync.ErrBlockIsNotSigned, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldWork(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}, process.BHProcessed)
	assert.Nil(t, err)
}

func TestBasicForkDetector_RemoveHeadersShouldWork(t *testing.T) {
	t.Parallel()
	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 2, Round: 2, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewShardForkDetector(rounderMock)

	rounderMock.RoundIndex = 1
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	rounderMock.RoundIndex = 2
	_ = bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))

	hInfos = bfd.GetHeaders(2)
	assert.Equal(t, 1, len(hInfos))

	bfd.RemoveHeaders(1, hash1)

	hInfos = bfd.GetHeaders(1)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(2)
	assert.Equal(t, 1, len(hInfos))
}

func TestBasicForkDetector_CheckForkOnlyOneShardHeaderOnANonceShouldReturnFalse(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 0, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 1, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHProcessed,
		nil,
		nil)
	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.False(t, forkDetected)
	assert.Equal(t, uint64(math.MaxUint64), lowestForkNonce)
	assert.Nil(t, forkHash)
}

func TestBasicForkDetector_CheckForkMetaHeaderNotProcessedYetShouldReturnFalse(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.False(t, forkDetected)
	assert.Equal(t, uint64(math.MaxUint64), lowestForkNonce)
	assert.Nil(t, forkHash)
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldReturnFalseWhenLowerRound(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)
	rounderMock.RoundIndex = 5
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)
	rounderMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	rounderMock.RoundIndex = 3
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.False(t, forkDetected)
	assert.Equal(t, uint64(math.MaxUint64), lowestForkNonce)
	assert.Nil(t, forkHash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldReturnFalseWhenEqualRoundWithLowerHash(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)
	rounderMock.RoundIndex = 5
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHReceived,
		nil,
		nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.False(t, forkDetected)
	assert.Equal(t, uint64(math.MaxUint64), lowestForkNonce)
	assert.Nil(t, forkHash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, []byte("hash1"), hInfos[0].Hash())
}

func TestBasicForkDetector_CheckForkShardHeaderProcessedShouldReturnTrueWhenEqualRoundWithLowerHash(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewShardForkDetector(rounderMock)

	hdr1 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")

	finalHeaders2 := []data.HeaderHandler{
		hdr2,
	}
	finalHeadersHashes2 := [][]byte{
		hash2,
	}
	finalHeaders3 := []data.HeaderHandler{
		hdr3,
	}
	finalHeadersHashes3 := [][]byte{
		hash3,
	}

	rounderMock.RoundIndex = 5
	_ = bfd.AddHeader(
		hdr1,
		hash1,
		process.BHProcessed,
		nil,
		nil)
	_ = bfd.AddHeader(
		hdr2,
		hash2,
		process.BHNotarized,
		finalHeaders2,
		finalHeadersHashes2)
	_ = bfd.AddHeader(
		hdr3,
		hash3,
		process.BHNotarized,
		finalHeaders3,
		finalHeadersHashes3)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.True(t, forkDetected)
	assert.Equal(t, uint64(1), lowestForkNonce)
	assert.Equal(t, hash2, forkHash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
	assert.Equal(t, hash1, hInfos[0].Hash())
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldReturnTrueWhenEqualRoundWithHigherHash(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)
	rounderMock.RoundIndex = 5
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHProcessed,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHReceived,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.True(t, forkDetected)
	assert.Equal(t, uint64(1), lowestForkNonce)
	assert.Equal(t, []byte("hash1"), forkHash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
}

func TestBasicForkDetector_CheckForkShardHeaderProcessedShouldReturnTrueWhenEqualRoundWithHigherHash(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewShardForkDetector(rounderMock)

	hdr1 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")

	finalHeaders1 := []data.HeaderHandler{
		hdr1,
	}
	finalHeadersHashes1 := [][]byte{
		hash1,
	}
	finalHeaders3 := []data.HeaderHandler{
		hdr3,
	}
	finalHeadersHashes3 := [][]byte{
		hash3,
	}

	rounderMock.RoundIndex = 5
	_ = bfd.AddHeader(
		hdr2,
		hash2,
		process.BHProcessed,
		nil,
		nil)
	_ = bfd.AddHeader(
		hdr3,
		hash3,
		process.BHNotarized,
		finalHeaders3,
		finalHeadersHashes3)
	_ = bfd.AddHeader(
		hdr1,
		hash1,
		process.BHNotarized,
		finalHeaders1,
		finalHeadersHashes1)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.True(t, forkDetected)
	assert.Equal(t, uint64(1), lowestForkNonce)
	assert.Equal(t, hash1, forkHash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
}

func TestBasicForkDetector_CheckForkShouldNotConsiderProposedBlocks(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	rounderMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	rounderMock.RoundIndex = 3
	_ = bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 2, PrevRandSeed: []byte("X"), RandSeed: []byte("X")},
		[]byte("hash2"),
		process.BHProposed,
		nil,
		nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 2, len(hInfos))

	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.False(t, forkDetected)
	assert.Equal(t, uint64(math.MaxUint64), lowestForkNonce)
	assert.Nil(t, forkHash)
}

func TestBasicForkDetector_CheckForkShouldReturnTrue(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)
	rounderMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)
	rounderMock.RoundIndex = 3
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	rounderMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkDetected, lowestForkNonce, forkHash := bfd.CheckFork()
	assert.True(t, forkDetected)
	assert.Equal(t, uint64(1), lowestForkNonce)
	assert.Equal(t, []byte("hash2"), forkHash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
}

func TestBasicForkDetector_RemovePastHeadersShouldWork(t *testing.T) {
	t.Parallel()
	hdr1 := &block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 2, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 3, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")
	rounderMock := &mock.RounderMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	_ = bfd.AddHeader(hdr1, hash1, process.BHReceived, nil, nil)
	_ = bfd.AddHeader(hdr2, hash2, process.BHReceived, nil, nil)
	_ = bfd.AddHeader(hdr3, hash3, process.BHReceived, nil, nil)
	bfd.SetFinalCheckpoint(4, 4)
	bfd.RemovePastHeaders()

	hInfos := bfd.GetHeaders(3)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(2)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(1)
	assert.Nil(t, hInfos)
}

func TestBasicForkDetector_RemoveInvalidReceivedHeadersShouldWork(t *testing.T) {
	t.Parallel()
	hdr0 := &block.Header{PubKeysBitmap: []byte("X"), Nonce: 8, Round: 10}
	hash0 := []byte("hash0")
	hdr1 := &block.Header{PubKeysBitmap: []byte("X"), Nonce: 9, Round: 12}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{PubKeysBitmap: []byte("X"), Nonce: 13, Round: 15}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{PubKeysBitmap: []byte("X"), Nonce: 10, Round: 14}
	hash3 := []byte("hash3")
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewShardForkDetector(rounderMock)
	rounderMock.RoundIndex = 11
	_ = bfd.AddHeader(hdr0, hash0, process.BHReceived, nil, nil)
	rounderMock.RoundIndex = 13
	_ = bfd.AddHeader(hdr1, hash1, process.BHReceived, nil, nil)
	rounderMock.RoundIndex = 16
	_ = bfd.AddHeader(hdr2, hash2, process.BHReceived, nil, nil)
	rounderMock.RoundIndex = 15
	_ = bfd.AddHeader(hdr3, hash3, process.BHReceived, nil, nil)
	bfd.SetFinalCheckpoint(9, 12)
	bfd.RemoveInvalidReceivedHeaders()

	hInfos := bfd.GetHeaders(8)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(9)
	assert.NotNil(t, hInfos)

	hInfos = bfd.GetHeaders(13)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(10)
	assert.NotNil(t, hInfos)
	assert.Equal(t, uint64(13), bfd.ProbableHighestNonce())
	assert.Equal(t, uint64(10), bfd.ComputeProbableHighestNonce())
}

func TestBasicForkDetector_RemoveCheckpointHeaderNonceShouldResetCheckpoint(t *testing.T) {
	t.Parallel()
	hdr1 := &block.Header{Nonce: 2, Round: 2, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	rounderMock := &mock.RounderMock{RoundIndex: 2}
	bfd, _ := sync.NewShardForkDetector(rounderMock)

	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(2), bfd.LastCheckpointNonce())

	bfd.RemoveHeaders(2, hash1)
	assert.Equal(t, uint64(0), bfd.LastCheckpointNonce())
	assert.Equal(t, uint64(0), bfd.LastCheckpointRound())
}

func TestBasicForkDetector_GetHighestFinalBlockNonce(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)

	hdr1 := &block.MetaBlock{Nonce: 2, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	rounderMock.RoundIndex = 1
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(0), bfd.GetHighestFinalBlockNonce())

	hdr2 := &block.MetaBlock{Nonce: 3, Round: 3, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	rounderMock.RoundIndex = 3
	_ = bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(0), bfd.GetHighestFinalBlockNonce())

	hdr3 := &block.MetaBlock{Nonce: 4, Round: 4, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")
	rounderMock.RoundIndex = 4
	_ = bfd.AddHeader(hdr3, hash3, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(3), bfd.GetHighestFinalBlockNonce())

	hdr4 := &block.MetaBlock{Nonce: 6, Round: 5, PubKeysBitmap: []byte("X")}
	hash4 := []byte("hash4")
	rounderMock.RoundIndex = 5
	_ = bfd.AddHeader(hdr4, hash4, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(3), bfd.GetHighestFinalBlockNonce())
}

func TestBasicForkDetector_ProbableHighestNonce(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)

	rounderMock.RoundIndex = 11
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 8, Round: 10},
		[]byte("hash0"),
		process.BHReceived,
		nil,
		nil)
	assert.Equal(t, uint64(8), bfd.ProbableHighestNonce())

	rounderMock.RoundIndex = 13
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 9, Round: 12},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	assert.Equal(t, uint64(9), bfd.ProbableHighestNonce())

	rounderMock.RoundIndex = 16
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 13, Round: 15},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	assert.Equal(t, uint64(13), bfd.ProbableHighestNonce())

	rounderMock.RoundIndex = 15
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 10, Round: 14},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)
	assert.Equal(t, uint64(10), bfd.ProbableHighestNonce())

	rounderMock.RoundIndex = 16
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 11, Round: 15},
		[]byte("hash3"),
		process.BHReceived,
		nil,
		nil)
	assert.Equal(t, uint64(11), bfd.ProbableHighestNonce())
}

func TestBasicForkDetector_ResetProbableHighestNonce(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)

	rounderMock.RoundIndex = 15
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 10, Round: 14},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)
	assert.Equal(t, uint64(10), bfd.ProbableHighestNonce())

	rounderMock.RoundIndex = 16
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 11, Round: 15},
		[]byte("hash3"),
		process.BHReceived,
		nil,
		nil)
	assert.Equal(t, uint64(11), bfd.ProbableHighestNonce())

	rounderMock.RoundIndex = 22
	bfd.ResetProbableHighestNonceIfNeeded()
	assert.Equal(t, uint64(10), bfd.ProbableHighestNonce())
}

func TestBasicForkDetector_GetProbableHighestNonce(t *testing.T) {
	rounderMock := &mock.RounderMock{}
	bfd, _ := sync.NewMetaForkDetector(rounderMock)

	hdr1 := &block.MetaBlock{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	rounderMock.RoundIndex = 2
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, uint64(1), bfd.GetProbableHighestNonce(hInfos))

	hdr2 := &block.MetaBlock{Nonce: 2, Round: 2, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	rounderMock.RoundIndex = 3
	_ = bfd.AddHeader(hdr2, hash2, process.BHReceived, nil, nil)
	hInfos = bfd.GetHeaders(2)
	assert.Equal(t, uint64(2), bfd.GetProbableHighestNonce(hInfos))

	hdr3 := &block.MetaBlock{Nonce: 3, Round: 3, PrevRandSeed: []byte("X"), RandSeed: []byte("X")}
	hash3 := []byte("hash3")
	rounderMock.RoundIndex = 4
	_ = bfd.AddHeader(hdr3, hash3, process.BHProposed, nil, nil)
	hInfos = bfd.GetHeaders(3)
	assert.Equal(t, uint64(2), bfd.GetProbableHighestNonce(hInfos))

	hdr4 := &block.MetaBlock{Nonce: 3, Round: 3, PubKeysBitmap: []byte("X")}
	hash4 := []byte("hash4")
	rounderMock.RoundIndex = 4
	_ = bfd.AddHeader(hdr4, hash4, process.BHReceived, nil, nil)
	hInfos = bfd.GetHeaders(3)
	assert.Equal(t, uint64(3), bfd.GetProbableHighestNonce(hInfos))
}

func TestShardForkDetector_ShouldAddBlockInForkDetectorShouldWork(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 10}
	sfd, _ := sync.NewShardForkDetector(rounderMock)

	hdr := &block.Header{Nonce: 1, Round: 1}
	err := sfd.ShouldAddBlockInForkDetector(hdr, process.BHProcessed, process.ShardBlockFinality)
	assert.Nil(t, err)

	sfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference + 1)
	err = sfd.ShouldAddBlockInForkDetector(hdr, process.BHReceived, process.ShardBlockFinality)
	assert.Nil(t, err)

	sfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference)
	hdr.Round = uint64(rounderMock.RoundIndex - process.ShardBlockFinality)
	err = sfd.ShouldAddBlockInForkDetector(hdr, process.BHReceived, process.ShardBlockFinality)
	assert.Nil(t, err)

	sfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference + 1)
	err = sfd.ShouldAddBlockInForkDetector(hdr, process.BHProposed, process.ShardBlockFinality)
	assert.Nil(t, err)
}

func TestShardForkDetector_ShouldAddBlockInForkDetectorShouldErrLowerRoundInBlock(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 10}
	sfd, _ := sync.NewShardForkDetector(rounderMock)
	hdr := &block.Header{Nonce: 1, Round: 1}

	hdr.Round = uint64(rounderMock.RoundIndex - process.ShardBlockFinality - 1)
	err := sfd.ShouldAddBlockInForkDetector(hdr, process.BHReceived, process.ShardBlockFinality)
	assert.Equal(t, sync.ErrLowerRoundInBlock, err)

	sfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference + 1)
	err = sfd.ShouldAddBlockInForkDetector(hdr, process.BHProposed, process.ShardBlockFinality)
	assert.Equal(t, sync.ErrLowerRoundInBlock, err)
}

func TestMetaForkDetector_ShouldAddBlockInForkDetectorShouldWork(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 10}
	mfd, _ := sync.NewMetaForkDetector(rounderMock)

	hdr := &block.MetaBlock{Nonce: 1, Round: 1}
	err := mfd.ShouldAddBlockInForkDetector(hdr, process.BHProcessed, process.MetaBlockFinality)
	assert.Nil(t, err)

	mfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference + 1)
	err = mfd.ShouldAddBlockInForkDetector(hdr, process.BHReceived, process.MetaBlockFinality)
	assert.Nil(t, err)

	mfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference)
	hdr.Round = uint64(rounderMock.RoundIndex - process.MetaBlockFinality)
	err = mfd.ShouldAddBlockInForkDetector(hdr, process.BHReceived, process.MetaBlockFinality)
	assert.Nil(t, err)

	mfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference + 1)
	err = mfd.ShouldAddBlockInForkDetector(hdr, process.BHProposed, process.MetaBlockFinality)
	assert.Nil(t, err)
}

func TestMetaForkDetector_ShouldAddBlockInForkDetectorShouldErrLowerRoundInBlock(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 10}
	mfd, _ := sync.NewMetaForkDetector(rounderMock)
	hdr := &block.MetaBlock{Nonce: 1, Round: 1}

	hdr.Round = uint64(rounderMock.RoundIndex - process.MetaBlockFinality - 1)
	err := mfd.ShouldAddBlockInForkDetector(hdr, process.BHReceived, process.MetaBlockFinality)
	assert.Equal(t, sync.ErrLowerRoundInBlock, err)

	mfd.SetProbableHighestNonce(hdr.GetNonce() + process.MaxNoncesDifference + 1)
	err = mfd.ShouldAddBlockInForkDetector(hdr, process.BHProposed, process.MetaBlockFinality)
	assert.Equal(t, sync.ErrLowerRoundInBlock, err)
}

func TestShardForkDetector_AddFinalHeadersShouldNotChangeTheFinalCheckpoint(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{RoundIndex: 10}
	sfd, _ := sync.NewShardForkDetector(rounderMock)
	hdr1 := &block.Header{Nonce: 3, Round: 3}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 4, Round: 5}
	hash3 := []byte("hash3")

	hdrs := make([]data.HeaderHandler, 0)
	hashes := make([][]byte, 0)
	hdrs = append(hdrs, hdr1)
	hashes = append(hashes, hash1)
	sfd.AddFinalHeaders(hdrs, hashes)
	assert.Equal(t, hdr1.Nonce, sfd.FinalCheckpointNonce())

	hdrs = make([]data.HeaderHandler, 0)
	hashes = make([][]byte, 0)
	hdrs = append(hdrs, hdr2)
	hashes = append(hashes, hash2)
	sfd.AddFinalHeaders(hdrs, hashes)
	assert.Equal(t, hdr1.Nonce, sfd.FinalCheckpointNonce())

	hdrs = make([]data.HeaderHandler, 0)
	hashes = make([][]byte, 0)
	hdrs = append(hdrs, hdr3)
	hashes = append(hashes, hash3)
	sfd.AddFinalHeaders(hdrs, hashes)
	assert.Equal(t, hdr3.Nonce, sfd.FinalCheckpointNonce())
}
