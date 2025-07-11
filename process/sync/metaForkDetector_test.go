package sync_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetaForkDetector_NilRoundHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		nil,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		nil,
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewMetaForkDetector_NilEnableRoundsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		nil,
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilEnableRoundsHandler, err)
}

func TestNewMetaForkDetector_NilProofsPoolShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewMetaForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		nil,
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
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

func TestMetaForkDetector_ComputeGenesisTimeFromHeader(t *testing.T) {
	t.Parallel()

	t.Run("legacy genesis time calculation", func(t *testing.T) {
		t.Parallel()

		roundDuration := uint64(100)
		roundHandlerMock := &mock.RoundHandlerMock{
			RoundTimeDuration: time.Duration(roundDuration) * time.Second,
		}

		genesisTime := int64(9000)
		hdrTimeStamp := uint64(10000)
		hdrRound := uint64(10)
		bfd, _ := sync.NewMetaForkDetector(
			roundHandlerMock,
			&testscommon.TimeCacheStub{},
			&mock.BlockTrackerMock{},
			genesisTime,
			0,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag != common.SupernovaFlag
				},
			},
			&testscommon.EnableRoundsHandlerStub{},
			&dataRetriever.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{},
		)

		hdr1 := &block.Header{Nonce: 1, Round: hdrRound, PubKeysBitmap: []byte("X"), TimeStamp: hdrTimeStamp}

		err := bfd.CheckGenesisTimeForHeader(hdr1)
		require.Nil(t, err)
	})

	t.Run("supernova activated in epoch but not in round", func(t *testing.T) {
		t.Parallel()

		roundDuration := uint64(100)
		roundHandlerMock := &mock.RoundHandlerMock{
			RoundTimeDuration: time.Duration(roundDuration) * time.Second,
		}

		genesisTime := int64(9000)
		hdrTimeStamp := uint64(10000000) // as milliseconds
		hdrRound := uint64(10)
		bfd, _ := sync.NewMetaForkDetector(
			roundHandlerMock,
			&testscommon.TimeCacheStub{},
			&mock.BlockTrackerMock{},
			genesisTime,
			0,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.SupernovaFlag && epoch > 0
				},
			},
			&testscommon.EnableRoundsHandlerStub{
				IsFlagEnabledCalled: func(flag common.EnableRoundFlag) bool {
					return flag != common.SupernovaRoundFlag
				},
			},
			&dataRetriever.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{},
		)

		hdr1 := &block.Header{
			Nonce:         1,
			Round:         hdrRound,
			Epoch:         1,
			PubKeysBitmap: []byte("X"),
			TimeStamp:     hdrTimeStamp,
		}

		err := bfd.CheckGenesisTimeForHeader(hdr1)
		assert.Nil(t, err)
	})

	t.Run("supernova activated in epoch and round", func(t *testing.T) {
		t.Parallel()

		roundDuration := uint64(1000)
		roundHandlerMock := &mock.RoundHandlerMock{
			RoundTimeDuration: time.Duration(roundDuration) * time.Millisecond,
		}

		genesisTime := int64(900)
		supernovaGenesisTime := int64(90000)
		hdrTimeStamp := uint64(100000) // as milliseconds

		hdrRound := uint64(20)
		supernovaActivationRound := uint64(10)

		bfd, _ := sync.NewMetaForkDetector(
			roundHandlerMock,
			&testscommon.TimeCacheStub{},
			&mock.BlockTrackerMock{},
			genesisTime,
			supernovaGenesisTime,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.SupernovaFlag
				},
			},
			&testscommon.EnableRoundsHandlerStub{
				IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
					return flag == common.SupernovaRoundFlag
				},
				GetActivationRoundCalled: func(flag common.EnableRoundFlag) uint64 {
					return supernovaActivationRound
				},
			},
			&dataRetriever.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						RoundDuration: roundDuration,
					}, nil
				},
			},
		)

		hdr1 := &block.Header{Nonce: 1, Round: hdrRound, PubKeysBitmap: []byte("X"), TimeStamp: hdrTimeStamp}

		err := bfd.CheckGenesisTimeForHeader(hdr1)
		require.Nil(t, err)
	})
}
