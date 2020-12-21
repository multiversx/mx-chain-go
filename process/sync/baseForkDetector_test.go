package sync_test

import (
	"math"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewBasicForkDetector_ShouldErrNilRoundHandler(t *testing.T) {
	t.Parallel()

	bfd, err := sync.NewShardForkDetector(
		nil,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	assert.Equal(t, process.ErrNilRoundHandler, err)
	assert.Nil(t, bfd)
}

func TestNewBasicForkDetector_ShouldErrNilBlackListHandler(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, err := sync.NewShardForkDetector(
		roundHandler,
		nil,
		&mock.BlockTrackerMock{},
		0,
	)
	assert.Equal(t, process.ErrNilBlackListCacher, err)
	assert.Nil(t, bfd)
}

func TestNewBasicForkDetector_ShouldErrNilBlockTracker(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, err := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		nil,
		0,
	)
	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, bfd)
}

func TestNewBasicForkDetector_ShouldWork(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, err := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	assert.Nil(t, err)
	assert.NotNil(t, bfd)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrGenesisTimeMissmatch(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now().Unix()
	roundTimeDuration := 4 * time.Second
	round := uint64(2)
	incorrectTimeStamp := uint64(genesisTime + int64(roundTimeDuration)*int64(round) - 1)

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1, RoundTimeDuration: roundTimeDuration}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		genesisTime,
	)

	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: round, TimeStamp: incorrectTimeStamp}, []byte("hash"))
	assert.Equal(t, sync.ErrGenesisTimeMissmatch, err)

	err = bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: round, TimeStamp: incorrectTimeStamp}, []byte("hash"))
	assert.Equal(t, sync.ErrGenesisTimeMissmatch, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrLowerRoundInBlock(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	bfd.SetFinalCheckpoint(1, 1, nil)
	err := bfd.CheckBlockValidity(&block.Header{PubKeysBitmap: []byte("X")}, []byte("hash"))
	assert.Equal(t, sync.ErrLowerRoundInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrLowerNonceInBlock(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	bfd.SetFinalCheckpoint(2, 2, nil)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")}, []byte("hash"))
	assert.Equal(t, sync.ErrLowerNonceInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrHigherRoundInBlock(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 0}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")}, []byte("hash"))
	assert.Equal(t, sync.ErrHigherRoundInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldErrHigherNonceInBlock(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 2, Round: 1, PubKeysBitmap: []byte("X")}, []byte("hash"))
	assert.Equal(t, sync.ErrHigherNonceInBlock, err)
}

func TestBasicForkDetector_CheckBlockValidityShouldWork(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	err := bfd.CheckBlockValidity(&block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}, []byte("hash"))
	assert.Nil(t, err)
}

func TestBasicForkDetector_RemoveHeadersShouldWork(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 2, Round: 2, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	roundHandlerMock.RoundIndex = 1
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	roundHandlerMock.RoundIndex = 2
	_ = bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))

	hInfos = bfd.GetHeaders(2)
	assert.Equal(t, 1, len(hInfos))

	bfd.RemoveHeader(1, hash1)

	hInfos = bfd.GetHeaders(1)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(2)
	assert.Equal(t, 1, len(hInfos))
}

func TestBasicForkDetector_CheckForkOnlyOneShardHeaderOnANonceShouldReturnFalse(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 99}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 0, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHProcessed,
		nil,
		nil)
	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Nonce)
	assert.Nil(t, forkInfo.Hash)
}

func TestBasicForkDetector_CheckForkOnlyReceivedHeadersShouldReturnFalse(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 99}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 0, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil,
	)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil,
	)
	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Nonce)
	assert.Nil(t, forkInfo.Hash)
}

func TestBasicForkDetector_CheckForkOnlyOneShardHeaderOnANonceReceivedAndProcessedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 99}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 0, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil,
	)
	_ = bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil,
	)
	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Nonce)
	assert.Nil(t, forkInfo.Hash)
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 99}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHProcessed,
		nil,
		nil)
	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Nonce)
	assert.Nil(t, forkInfo.Hash)
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldReturnFalseWhenLowerRound(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 5
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 3
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Nonce)
	assert.Nil(t, forkInfo.Hash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldReturnFalseWhenEqualRoundWithLowerHash(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 5
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

	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Nonce)
	assert.Nil(t, forkInfo.Hash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
	assert.Equal(t, []byte("hash1"), hInfos[0].Hash())
}

func TestBasicForkDetector_CheckForkShardHeaderProcessedShouldReturnTrueWhenEqualRoundWithLowerHash(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	hdr1 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")

	selfNotarizedHeaders2 := []data.HeaderHandler{
		hdr2,
	}
	selfNotarizedHeadersHashes2 := [][]byte{
		hash2,
	}
	selfNotarizedHeaders3 := []data.HeaderHandler{
		hdr3,
	}
	selfNotarizedHeadersHashes3 := [][]byte{
		hash3,
	}

	roundHandlerMock.RoundIndex = 5
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
		selfNotarizedHeaders2,
		selfNotarizedHeadersHashes2)
	_ = bfd.AddHeader(
		hdr3,
		hash3,
		process.BHNotarized,
		selfNotarizedHeaders3,
		selfNotarizedHeadersHashes3)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkInfo := bfd.CheckFork()
	assert.True(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(1), forkInfo.Nonce)
	assert.Equal(t, hash2, forkInfo.Hash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
	assert.Equal(t, hash1, hInfos[0].Hash())
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldReturnTrueWhenEqualRoundWithHigherHash(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 5
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

	forkInfo := bfd.CheckFork()
	assert.True(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(1), forkInfo.Nonce)
	assert.Equal(t, []byte("hash1"), forkInfo.Hash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
}

func TestBasicForkDetector_CheckForkShardHeaderProcessedShouldReturnTrueWhenEqualRoundWithHigherHash(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	hdr1 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")

	selfNotarizedHeaders1 := []data.HeaderHandler{
		hdr1,
	}
	selfNotarizedHeadersHashes1 := [][]byte{
		hash1,
	}
	selfNotarizedHeaders3 := []data.HeaderHandler{
		hdr3,
	}
	selfNotarizedHeadersHashes3 := [][]byte{
		hash3,
	}

	roundHandlerMock.RoundIndex = 5
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
		selfNotarizedHeaders3,
		selfNotarizedHeadersHashes3)
	_ = bfd.AddHeader(
		hdr1,
		hash1,
		process.BHNotarized,
		selfNotarizedHeaders1,
		selfNotarizedHeadersHashes1)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkInfo := bfd.CheckFork()
	assert.True(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(1), forkInfo.Nonce)
	assert.Equal(t, hash1, forkInfo.Hash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
}

func TestBasicForkDetector_CheckForkShouldReturnTrue(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 3
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))

	forkInfo := bfd.CheckFork()
	assert.True(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(1), forkInfo.Nonce)
	assert.Equal(t, []byte("hash2"), forkInfo.Hash)

	hInfos = bfd.GetHeaders(1)
	assert.Equal(t, 3, len(hInfos))
}

func TestBasicForkDetector_CheckForkShouldReturnFalseWhenForkIsOnFinalCheckpointNonce(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 1
	_ = bfd.AddHeader(
		&block.MetaBlock{Epoch: 0, Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 4
	_ = bfd.AddHeader(
		&block.MetaBlock{Epoch: 1, Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 5
	_ = bfd.AddHeader(
		&block.MetaBlock{Epoch: 1, Nonce: 2, Round: 5, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)

	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
}

func TestBasicForkDetector_CheckForkShouldReturnFalseWhenForkIsOnHigherEpochBlockReceivedTooLate(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 2
	_ = bfd.AddHeader(
		&block.MetaBlock{Epoch: 0, Nonce: 1, Round: 2, PubKeysBitmap: []byte("X")},
		[]byte("hash2"),
		process.BHProcessed,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 3
	_ = bfd.AddHeader(
		&block.MetaBlock{Epoch: 0, Nonce: 2, Round: 3, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
		process.BHProposed,
		nil,
		nil)
	roundHandlerMock.RoundIndex = 3
	_ = bfd.AddHeader(
		&block.MetaBlock{Epoch: 1, Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)

	forkInfo := bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
}

func TestBasicForkDetector_RemovePastHeadersShouldWork(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 2, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 3, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	_ = bfd.AddHeader(hdr1, hash1, process.BHReceived, nil, nil)
	_ = bfd.AddHeader(hdr2, hash2, process.BHReceived, nil, nil)
	_ = bfd.AddHeader(hdr3, hash3, process.BHReceived, nil, nil)
	bfd.SetFinalCheckpoint(4, 4, nil)
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
	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 11
	_ = bfd.AddHeader(hdr0, hash0, process.BHReceived, nil, nil)
	roundHandlerMock.RoundIndex = 13
	_ = bfd.AddHeader(hdr1, hash1, process.BHReceived, nil, nil)
	roundHandlerMock.RoundIndex = 16
	_ = bfd.AddHeader(hdr2, hash2, process.BHReceived, nil, nil)
	roundHandlerMock.RoundIndex = 15
	_ = bfd.AddHeader(hdr3, hash3, process.BHReceived, nil, nil)
	bfd.SetFinalCheckpoint(9, 12, nil)
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
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 2}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(2), bfd.LastCheckpointNonce())

	bfd.RemoveHeader(2, hash1)
	assert.Equal(t, uint64(0), bfd.LastCheckpointNonce())
	assert.Equal(t, uint64(0), bfd.LastCheckpointRound())
}

func TestBasicForkDetector_GetHighestFinalBlockNonce(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	hdr1 := &block.MetaBlock{Nonce: 2, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	roundHandlerMock.RoundIndex = 1
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(0), bfd.GetHighestFinalBlockNonce())

	hdr2 := &block.MetaBlock{Nonce: 3, Round: 3, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	roundHandlerMock.RoundIndex = 3
	_ = bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(0), bfd.GetHighestFinalBlockNonce())

	hdr3 := &block.MetaBlock{Nonce: 4, Round: 4, PubKeysBitmap: []byte("X")}
	hash3 := []byte("hash3")
	roundHandlerMock.RoundIndex = 4
	_ = bfd.AddHeader(hdr3, hash3, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(3), bfd.GetHighestFinalBlockNonce())

	hdr4 := &block.MetaBlock{Nonce: 6, Round: 5, PubKeysBitmap: []byte("X")}
	hash4 := []byte("hash4")
	roundHandlerMock.RoundIndex = 5
	_ = bfd.AddHeader(hdr4, hash4, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(3), bfd.GetHighestFinalBlockNonce())
}

func TestBasicForkDetector_ProbableHighestNonce(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	roundHandlerMock.RoundIndex = 11
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 8, Round: 10},
		[]byte("hash0"),
		process.BHReceived,
		nil,
		nil)
	assert.Equal(t, uint64(8), bfd.ProbableHighestNonce())

	roundHandlerMock.RoundIndex = 13
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 9, Round: 12},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil)
	assert.Equal(t, uint64(9), bfd.ProbableHighestNonce())

	roundHandlerMock.RoundIndex = 16
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 13, Round: 15},
		[]byte("hash2"),
		process.BHReceived,
		nil,
		nil)
	assert.Equal(t, uint64(13), bfd.ProbableHighestNonce())

	roundHandlerMock.RoundIndex = 15
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 10, Round: 14},
		[]byte("hash3"),
		process.BHProcessed,
		nil,
		nil)
	assert.Equal(t, uint64(10), bfd.ProbableHighestNonce())

	roundHandlerMock.RoundIndex = 16
	_ = bfd.AddHeader(
		&block.MetaBlock{PubKeysBitmap: []byte("X"), Nonce: 11, Round: 15},
		[]byte("hash3"),
		process.BHReceived,
		nil,
		nil)
	assert.Equal(t, uint64(11), bfd.ProbableHighestNonce())
}

func TestShardForkDetector_ShouldAddBlockInForkDetectorShouldWork(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 10}
	sfd, _ := sync.NewShardForkDetector(roundHandlerMock, &mock.BlackListHandlerStub{}, &mock.BlockTrackerMock{}, 0)

	hdr := &block.Header{Nonce: 1, Round: 1}
	receivedTooLate := sfd.IsHeaderReceivedTooLate(hdr, process.BHProcessed, process.BlockFinality)
	assert.False(t, receivedTooLate)

	receivedTooLate = sfd.IsHeaderReceivedTooLate(hdr, process.BHReceived, process.BlockFinality)
	assert.True(t, receivedTooLate)

	hdr.Round = uint64(roundHandlerMock.RoundIndex - process.BlockFinality)
	receivedTooLate = sfd.IsHeaderReceivedTooLate(hdr, process.BHReceived, process.BlockFinality)
	assert.False(t, receivedTooLate)
}

func TestShardForkDetector_ShouldAddBlockInForkDetectorShouldErrLowerRoundInBlock(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 10}
	sfd, _ := sync.NewShardForkDetector(roundHandlerMock, &mock.BlackListHandlerStub{}, &mock.BlockTrackerMock{}, 0)
	hdr := &block.Header{Nonce: 1, Round: 1}

	hdr.Round = uint64(roundHandlerMock.RoundIndex - process.BlockFinality - 1)
	receivedTooLate := sfd.IsHeaderReceivedTooLate(hdr, process.BHReceived, process.BlockFinality)
	assert.True(t, receivedTooLate)
}

func TestMetaForkDetector_ShouldAddBlockInForkDetectorShouldWork(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 10}
	mfd, _ := sync.NewMetaForkDetector(roundHandlerMock, &mock.BlackListHandlerStub{}, &mock.BlockTrackerMock{}, 0)

	hdr := &block.MetaBlock{Nonce: 1, Round: 1}
	receivedTooLate := mfd.IsHeaderReceivedTooLate(hdr, process.BHProcessed, process.BlockFinality)
	assert.False(t, receivedTooLate)

	receivedTooLate = mfd.IsHeaderReceivedTooLate(hdr, process.BHReceived, process.BlockFinality)
	assert.True(t, receivedTooLate)

	hdr.Round = uint64(roundHandlerMock.RoundIndex - process.BlockFinality)
	receivedTooLate = mfd.IsHeaderReceivedTooLate(hdr, process.BHReceived, process.BlockFinality)
	assert.False(t, receivedTooLate)
}

func TestMetaForkDetector_ShouldAddBlockInForkDetectorShouldErrLowerRoundInBlock(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 10}
	mfd, _ := sync.NewMetaForkDetector(roundHandlerMock, &mock.BlackListHandlerStub{}, &mock.BlockTrackerMock{}, 0)
	hdr := &block.MetaBlock{Nonce: 1, Round: 1}

	hdr.Round = uint64(roundHandlerMock.RoundIndex - process.BlockFinality - 1)
	receivedTooLate := mfd.IsHeaderReceivedTooLate(hdr, process.BHReceived, process.BlockFinality)
	assert.True(t, receivedTooLate)
}

func TestShardForkDetector_AddNotarizedHeadersShouldNotChangeTheFinalCheckpoint(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 10}
	sfd, _ := sync.NewShardForkDetector(roundHandlerMock, &mock.BlackListHandlerStub{}, &mock.BlockTrackerMock{}, 0)
	hdr1 := &block.Header{Nonce: 3, Round: 3}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 4, Round: 4}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 5, Round: 5}
	hash3 := []byte("hash3")

	hdrs := make([]data.HeaderHandler, 0)
	hashes := make([][]byte, 0)
	hdrs = append(hdrs, hdr1)
	hashes = append(hashes, hash1)

	sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, hdrs, hashes)
	assert.Equal(t, uint64(0), sfd.FinalCheckpointNonce())

	_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, hdrs, hashes)
	assert.Equal(t, hdr1.Nonce, sfd.FinalCheckpointNonce())

	hdrs = make([]data.HeaderHandler, 0)
	hashes = make([][]byte, 0)
	hdrs = append(hdrs, hdr2)
	hashes = append(hashes, hash2)

	sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, hdrs, hashes)
	assert.Equal(t, hdr1.Nonce, sfd.FinalCheckpointNonce())

	_ = sfd.AddHeader(hdr2, hash2, process.BHProcessed, hdrs, hashes)
	assert.Equal(t, hdr2.Nonce, sfd.FinalCheckpointNonce())

	hdrs = make([]data.HeaderHandler, 0)
	hashes = make([][]byte, 0)
	hdrs = append(hdrs, hdr3)
	hashes = append(hashes, hash3)

	sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, hdrs, hashes)
	assert.Equal(t, hdr2.Nonce, sfd.FinalCheckpointNonce())

	_ = sfd.AddHeader(hdr3, hash3, process.BHProcessed, hdrs, hashes)
	assert.Equal(t, hdr3.Nonce, sfd.FinalCheckpointNonce())
}

func TestBaseForkDetector_IsConsensusStuckNotSyncingShouldReturnFalse(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewShardForkDetector(roundHandlerMock, &mock.BlackListHandlerStub{}, &mock.BlockTrackerMock{}, 0)

	bfd.SetProbableHighestNonce(1)

	assert.False(t, bfd.IsConsensusStuck())
}

func TestBaseForkDetector_IsConsensusStuckNoncesDifferencesNotEnoughShouldReturnFalse(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	roundHandlerMock.RoundIndex = 10
	assert.False(t, bfd.IsConsensusStuck())
}

func TestBaseForkDetector_IsConsensusStuckNotInProperRoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	roundHandlerMock.RoundIndex = 11
	assert.False(t, bfd.IsConsensusStuck())
}

func TestBaseForkDetector_IsConsensusStuckShouldReturnTrue(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	// last checkpoint will be (round = 0 , nonce = 0)
	// round difference is higher than 10
	// round index is divisible by RoundModulusTrigger -> 5
	// => consensus is stuck
	roundHandlerMock.RoundIndex = 20
	assert.True(t, bfd.IsConsensusStuck())
}

func TestBaseForkDetector_ComputeTimeDuration(t *testing.T) {
	t.Parallel()

	roundDuration := uint64(1)
	roundHandlerMock := &mock.RoundHandlerMock{
		RoundTimeDuration: time.Second,
	}

	genesisTime := int64(9000)
	hdrTimeStamp := uint64(10000)
	hdrRound := uint64(20)
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		genesisTime,
	)

	hdr1 := &block.Header{Nonce: 1, Round: hdrRound, PubKeysBitmap: []byte("X"), TimeStamp: hdrTimeStamp}

	expectedTimeStamp := hdrTimeStamp - (hdrRound * roundDuration)
	timeDuration := bfd.ComputeGenesisTimeFromHeader(hdr1)
	assert.Equal(t, int64(expectedTimeStamp), timeDuration)
}

func TestShardForkDetector_RemoveHeaderShouldComputeFinalCheckpoint(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 10}
	sfd, _ := sync.NewShardForkDetector(roundHandlerMock, &mock.BlackListHandlerStub{}, &mock.BlockTrackerMock{}, 0)
	hdr1 := &block.Header{Nonce: 3, Round: 3}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 4, Round: 4}
	hash2 := []byte("hash2")

	hdrs := make([]data.HeaderHandler, 0)
	hashes := make([][]byte, 0)
	hdrs = append(hdrs, hdr1)
	hashes = append(hashes, hash1)

	_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, uint64(0), sfd.FinalCheckpointNonce())

	sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, hdrs, hashes)
	assert.Equal(t, hdr1.Nonce, sfd.FinalCheckpointNonce())

	hdrs = make([]data.HeaderHandler, 0)
	hashes = make([][]byte, 0)
	hdrs = append(hdrs, hdr2)
	hashes = append(hashes, hash2)

	_ = sfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)
	assert.Equal(t, hdr1.Nonce, sfd.FinalCheckpointNonce())

	sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, hdrs, hashes)
	assert.Equal(t, hdr2.Nonce, sfd.FinalCheckpointNonce())

	sfd.RemoveHeader(hdr2.GetNonce(), hash2)
	assert.Equal(t, hdr1.Nonce, sfd.FinalCheckpointNonce())
}

func TestBasicForkDetector_CheckForkMetaHeaderProcessedShouldWorkOnEqualRoundWithLowerHash(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)
	roundHandlerMock.RoundIndex = 5
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 1, Round: 4, PubKeysBitmap: []byte("X")},
		[]byte("hash3"),
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
		[]byte("hash1"),
		process.BHReceived,
		nil,
		nil)

	forkInfo := bfd.CheckFork()
	assert.True(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(1), forkInfo.Nonce)
	assert.Equal(t, uint64(4), forkInfo.Round)
	assert.Equal(t, []byte("hash1"), forkInfo.Hash)

	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 2, Round: 5, PubKeysBitmap: []byte("X")},
		[]byte("hash"),
		process.BHProposed,
		nil,
		nil)

	forkInfo = bfd.CheckFork()
	assert.False(t, forkInfo.IsDetected)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Nonce)
	assert.Equal(t, uint64(math.MaxUint64), forkInfo.Round)
	assert.Nil(t, forkInfo.Hash)
}

func TestBasicForkDetector_SetFinalToLastCheckpointShouldWork(t *testing.T) {
	roundHandlerMock := &mock.RoundHandlerMock{}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&mock.BlackListHandlerStub{},
		&mock.BlockTrackerMock{},
		0,
	)

	roundHandlerMock.RoundIndex = 1000
	_ = bfd.AddHeader(
		&block.MetaBlock{Nonce: 900, Round: 1000, PubKeysBitmap: []byte("X")},
		[]byte("hash"),
		process.BHProcessed,
		nil,
		nil)

	assert.Equal(t, uint64(0), bfd.GetHighestFinalBlockNonce())
	assert.Nil(t, bfd.GetHighestFinalBlockHash())

	bfd.SetFinalToLastCheckpoint()

	assert.Equal(t, uint64(900), bfd.GetHighestFinalBlockNonce())
	assert.Equal(t, []byte("hash"), bfd.GetHighestFinalBlockHash())
}
