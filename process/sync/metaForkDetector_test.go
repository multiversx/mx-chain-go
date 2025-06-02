package sync_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"

	"github.com/stretchr/testify/assert"
)

func TestNewMetaForkDetector_NilRoundHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		nil,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilRoundHandler, err)
}

func TestNewMetaForkDetector_NilBlackListShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		&mock.RoundHandlerMock{},
		nil,
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewMetaForkDetector_NilBlockTrackerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		nil,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewMetaForkDetector_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		nil,
		&dataRetriever.ProofsPoolMock{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewMetaForkDetector_NilProofsPoolShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		nil,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewMetaForkDetector_OkParamsShouldWork(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(sfd))

	assert.Equal(t, uint64(0), sfd.LastCheckpointNonce())
	assert.Equal(t, uint64(0), sfd.LastCheckpointRound())
	assert.Equal(t, uint64(0), sfd.FinalCheckpointNonce())
	assert.Equal(t, uint64(0), sfd.FinalCheckpointRound())
}

func TestMetaForkDetector_AddHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	err := bfd.AddHeader(nil, make([]byte, 0), process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHeader, err)
}

func TestMetaForkDetector_AddHeaderNilHashShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	err := bfd.AddHeader(&block.Header{}, nil, process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHash, err)
}

func TestMetaForkDetector_AddHeaderNotPresentShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := make([]byte, 0)
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	err := bfd.AddHeader(hdr, hash, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, hash, hInfos[0].Hash())
}

func TestMetaForkDetector_AddHeaderPresentShouldAppend(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	err := bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 2, len(hInfos))
	assert.Equal(t, hash1, hInfos[0].Hash())
	assert.Equal(t, hash2, hInfos[1].Hash())
}

func TestMetaForkDetector_AddHeaderWithProcessedBlockShouldSetCheckpoint(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 69, Round: 72, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 73}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, hdr1.Nonce, bfd.LastCheckpointNonce())
}

func TestMetaForkDetector_AddHeaderWithProcessedBlockAndFlagShouldSetCheckpoint(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 23, Round: 25, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 26}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		},
		&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		},
	)
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, hdr1.Nonce, bfd.FinalCheckpointNonce())
}

func TestMetaForkDetector_AddHeaderPresentShouldNotRewriteState(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	_ = bfd.AddHeader(hdr1, hash, process.BHReceived, nil, nil)
	err := bfd.AddHeader(hdr2, hash, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 2, len(hInfos))
	assert.Equal(t, hash, hInfos[0].Hash())
	assert.Equal(t, process.BHReceived, hInfos[0].GetBlockHeaderState())
	assert.Equal(t, process.BHProcessed, hInfos[1].GetBlockHeaderState())
}

func TestMetaForkDetector_AddHeaderHigherNonceThanRoundShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewMetaForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)
	err := bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 0, PubKeysBitmap: []byte("X")},
		[]byte("hash1"),
		process.BHProcessed,
		nil,
		nil,
	)
	assert.Equal(t, sync.ErrHigherNonceInBlock, err)
}
