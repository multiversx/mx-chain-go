package sync_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type blockTrackerWithInvalidationMock struct {
	*mock.BlockTrackerMock
	registerInvalidation func(handler func(uint32, []data.HeaderHandler, [][]byte))
}

func (tracker *blockTrackerWithInvalidationMock) RegisterInvalidatedSelfNotarizedFromCrossHeadersHandler(
	handler func(uint32, []data.HeaderHandler, [][]byte),
) {
	tracker.registerInvalidation(handler)
}

func TestNewShardForkDetector_NilRoundHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		nil,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilRoundHandler, err)
}

func TestNewShardForkDetector_NilBlackListShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{},
		nil,
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewShardForkDetector_NilBlockTrackerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		nil,
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewShardForkDetector_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		nil,
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewShardForkDetector_NilEnableRoundsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		nil,
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilEnableRoundsHandler, err)
}

func TestNewShardForkDetector_NilProofsPoolShouldErr(t *testing.T) {
	t.Parallel()

	sfd, err := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		nil,
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	assert.True(t, check.IfNil(sfd))
	assert.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewShardForkDetector_OkParamsShouldWork(t *testing.T) {
	t.Parallel()

	var invalidationHandler func(uint32, []data.HeaderHandler, [][]byte)
	blockTracker := &blockTrackerWithInvalidationMock{
		BlockTrackerMock: &mock.BlockTrackerMock{},
		registerInvalidation: func(handler func(uint32, []data.HeaderHandler, [][]byte)) {
			invalidationHandler = handler
		},
	}
	sfd, err := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{},
		&testscommon.TimeCacheStub{},
		blockTracker,
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(sfd))

	assert.Equal(t, uint64(0), sfd.LastCheckpointNonce())
	assert.Equal(t, uint64(0), sfd.LastCheckpointRound())
	assert.Equal(t, uint64(0), sfd.FinalCheckpointNonce())
	assert.Equal(t, uint64(0), sfd.FinalCheckpointRound())
	assert.NotNil(t, invalidationHandler)
}

func TestShardForkDetector_AddHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	err := bfd.AddHeader(nil, make([]byte, 0), process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHeader, err)
}

func TestShardForkDetector_AddHeaderNilHashShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	err := bfd.AddHeader(&block.Header{}, nil, process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrNilHash, err)
}

func TestShardForkDetector_AddHeaderNotPresentShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := make([]byte, 0)
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	err := bfd.AddHeader(hdr, hash, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, hash, hInfos[0].Hash())
}

func TestShardForkDetector_AddHeaderPresentShouldAppend(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash2 := []byte("hash2")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	err := bfd.AddHeader(hdr2, hash2, process.BHProcessed, nil, nil)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(1)
	assert.Equal(t, 2, len(hInfos))
	assert.Equal(t, hash1, hInfos[0].Hash())
	assert.Equal(t, hash2, hInfos[1].Hash())
}

func TestShardForkDetector_AddHeaderWithProcessedBlockShouldSetCheckpoint(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 69, Round: 72, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 73}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	_ = bfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
	assert.Equal(t, hdr1.Nonce, bfd.LastCheckpointNonce())
}

func TestShardForkDetector_AddHeaderPresentShouldNotRewriteState(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	hash := []byte("hash1")
	hdr2 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 1}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
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

func TestShardForkDetector_AddHeaderHigherNonceThanRoundShouldErr(t *testing.T) {
	t.Parallel()

	roundHandlerMock := &mock.RoundHandlerMock{RoundIndex: 100}
	bfd, _ := sync.NewShardForkDetector(
		roundHandlerMock,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)
	err := bfd.AddHeader(
		&block.Header{Nonce: 1, Round: 0, PubKeysBitmap: []byte("X")}, []byte("hash1"), process.BHProcessed, nil, nil)
	assert.Equal(t, sync.ErrHigherNonceInBlock, err)
}

func TestShardForkDetector_ComputeGenesisTimeFromHeader(t *testing.T) {
	t.Parallel()

	t.Run("legacy genesis time calculation", func(t *testing.T) {
		t.Parallel()

		roundDuration := uint64(100)
		roundHandlerMock := &mock.RoundHandlerMock{}

		genesisTime := int64(9000)
		hdrTimeStamp := uint64(10000)
		hdrRound := uint64(10)
		bfd, _ := sync.NewShardForkDetector(
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
			&dataRetrieverMock.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						RoundDuration: roundDuration * 1000,
					}, nil
				},
			},
			testscommon.GetDefaultProcessConfigsHandler(),
			0,
		)

		hdr1 := &block.Header{Nonce: 1, Round: hdrRound, PubKeysBitmap: []byte("X"), TimeStamp: hdrTimeStamp}

		err := bfd.CheckGenesisTimeForHeader(hdr1)
		require.Nil(t, err)
	})

	t.Run("legacy genesis time calculation, should fail if not able to get round duration", func(t *testing.T) {
		t.Parallel()

		expErr := errors.New("expected err")

		genesisTime := int64(9000)
		hdrTimeStamp := uint64(10000)
		hdrRound := uint64(10)
		bfd, _ := sync.NewMetaForkDetector(
			&mock.RoundHandlerMock{},
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
			&dataRetrieverMock.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{}, expErr
				},
			},
			testscommon.GetDefaultProcessConfigsHandler(),
		)

		hdr1 := &block.Header{Nonce: 1, Round: hdrRound, PubKeysBitmap: []byte("X"), TimeStamp: hdrTimeStamp}

		err := bfd.CheckGenesisTimeForHeader(hdr1)
		require.Equal(t, expErr, err)
	})

	t.Run("supernova activated in epoch but not in round", func(t *testing.T) {
		t.Parallel()

		roundDuration := uint64(100)
		roundHandlerMock := &mock.RoundHandlerMock{}

		genesisTime := int64(9000)
		hdrTimeStamp := uint64(10000000) // as milliseconds
		hdrRound := uint64(10)
		bfd, _ := sync.NewShardForkDetector(
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
			&dataRetrieverMock.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						RoundDuration: roundDuration * 1000,
					}, nil
				},
			},
			testscommon.GetDefaultProcessConfigsHandler(),
			0,
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

	t.Run("supernova activated in epoch but not in round, should fail if not able to get round duration", func(t *testing.T) {
		t.Parallel()

		expErr := errors.New("expected err")

		genesisTime := int64(9000)
		hdrTimeStamp := uint64(10000)
		hdrRound := uint64(10)
		bfd, _ := sync.NewShardForkDetector(
			&mock.RoundHandlerMock{},
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
			&dataRetrieverMock.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{}, expErr
				},
			},
			testscommon.GetDefaultProcessConfigsHandler(),
			0,
		)

		hdr1 := &block.Header{Nonce: 1, Round: hdrRound, Epoch: 2, PubKeysBitmap: []byte("X"), TimeStamp: hdrTimeStamp}

		err := bfd.CheckGenesisTimeForHeader(hdr1)
		require.Equal(t, expErr, err)
	})

	t.Run("supernova activated in epoch and round", func(t *testing.T) {
		t.Parallel()

		roundDuration := uint64(1000)
		roundHandlerMock := &mock.RoundHandlerMock{}

		genesisTime := int64(900)
		supernovaGenesisTime := int64(90000)
		hdrTimeStamp := uint64(100000) // as milliseconds

		hdrRound := uint64(20)
		supernovaActivationRound := uint64(10)

		bfd, _ := sync.NewShardForkDetector(
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
			&dataRetrieverMock.ProofsPoolMock{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						RoundDuration: roundDuration,
					}, nil
				},
			},
			testscommon.GetDefaultProcessConfigsHandler(),
			0,
		)

		hdr1 := &block.Header{Nonce: 1, Round: hdrRound, PubKeysBitmap: []byte("X"), TimeStamp: hdrTimeStamp}

		err := bfd.CheckGenesisTimeForHeader(hdr1)
		require.Nil(t, err)
	})
}

func createShardForkDetectorForFinality(enableEpochsHandler common.EnableEpochsHandler) interface {
	process.ForkDetector
	FinalCheckpointNonce() uint64
	SettledCheckpointNonce() uint64
	ReceivedSelfNotarizedFromCrossHeaders(shardID uint32, headers []data.HeaderHandler, hashes [][]byte)
	InvalidatedSelfNotarizedFromCrossHeaders(shardID uint32, headers []data.HeaderHandler, hashes [][]byte)
} {
	sfd, _ := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{RoundIndex: 5},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		enableEpochsHandler,
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)

	return sfd
}

func createShardForkDetectorForV3Settlement(enableEpochsHandler common.EnableEpochsHandler) interface {
	process.ForkDetector
	FinalCheckpointNonce() uint64
	SettledCheckpointNonce() uint64
	ReceivedSelfNotarizedFromCrossHeaders(shardID uint32, headers []data.HeaderHandler, hashes [][]byte)
	InvalidatedSelfNotarizedFromCrossHeaders(shardID uint32, headers []data.HeaderHandler, hashes [][]byte)
} {
	sfd, _ := sync.NewShardForkDetector(
		&mock.RoundHandlerMock{RoundIndex: 5},
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		enableEpochsHandler,
		testscommon.NewEnableRoundsHandlerStub(common.SupernovaRoundFlag),
		&dataRetrieverMock.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)

	return sfd
}

func TestShardForkDetector_DeferredFinalityUnderSupernova(t *testing.T) {
	t.Parallel()

	supernovaHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag || flag == common.SupernovaFlag
		},
	}
	andromedaOnlyHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag
		},
	}

	hash1, hash2, hash3 := []byte("hash1"), []byte("hash2"), []byte("hash3")
	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	contendedHdr2 := &block.Header{Nonce: 2, Round: 4, PrevHash: hash1, PubKeysBitmap: []byte("X")}
	cleanHdr3 := &block.Header{Nonce: 3, Round: 5, PrevHash: hash2, PubKeysBitmap: []byte("X")}

	t.Run("contended block defers finality, settles on notarization and cascades over descendants", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(supernovaHandler)

		_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(1), sfd.FinalCheckpointNonce())

		_ = sfd.AddHeader(contendedHdr2, hash2, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(1), sfd.FinalCheckpointNonce())

		// clean descendant of an unsettled block is not final either
		_ = sfd.AddHeader(cleanHdr3, hash3, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(1), sfd.FinalCheckpointNonce())

		// meta notarization settles the contended block; the clean descendant finalizes with it
		sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, []data.HeaderHandler{contendedHdr2}, [][]byte{hash2})
		require.Equal(t, uint64(3), sfd.FinalCheckpointNonce())
	})

	t.Run("non-contended blocks keep instant finality", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(supernovaHandler)

		_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(1), sfd.FinalCheckpointNonce())

		cleanHdr2 := &block.Header{Nonce: 2, Round: 2, PrevHash: hash1, PubKeysBitmap: []byte("X")}
		_ = sfd.AddHeader(cleanHdr2, hash2, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(2), sfd.FinalCheckpointNonce())
	})

	t.Run("notarization before processing advances the same final and settled checkpoint", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(supernovaHandler)
		sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, []data.HeaderHandler{hdr1}, [][]byte{hash1})
		require.Equal(t, uint64(0), sfd.FinalCheckpointNonce())
		require.Equal(t, uint64(0), sfd.SettledCheckpointNonce())

		_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(1), sfd.FinalCheckpointNonce())
		require.Equal(t, uint64(1), sfd.SettledCheckpointNonce())
	})

	t.Run("fork is signaled at the deferred nonce", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(supernovaHandler)

		_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
		_ = sfd.AddHeader(contendedHdr2, hash2, process.BHProcessed, nil, nil)

		sfd.ReceivedProof(&block.HeaderProof{
			HeaderHash:    []byte("competitorHash"),
			HeaderNonce:   2,
			HeaderRound:   3,
			HeaderShardId: 0,
		})

		forkInfo := sfd.CheckFork()
		require.True(t, forkInfo.IsDetected)
		require.Equal(t, uint64(2), forkInfo.Nonce)
	})

	t.Run("andromeda is instantly final", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(andromedaOnlyHandler)

		_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
		_ = sfd.AddHeader(contendedHdr2, hash2, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(2), sfd.FinalCheckpointNonce())
	})
}

func TestShardForkDetector_SettledWatermarkUnderSupernova(t *testing.T) {
	t.Parallel()

	supernovaHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag || flag == common.SupernovaFlag
		},
	}
	andromedaOnlyHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag
		},
	}

	hash1, hash2 := []byte("hash1"), []byte("hash2")
	hdr1 := &block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
	cleanHdr2 := &block.Header{Nonce: 2, Round: 2, PrevHash: hash1, PubKeysBitmap: []byte("X")}

	t.Run("instant finality does not advance the settled watermark, meta notarization does", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(supernovaHandler)

		_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
		_ = sfd.AddHeader(cleanHdr2, hash2, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(2), sfd.FinalCheckpointNonce())
		require.Equal(t, uint64(0), sfd.SettledCheckpointNonce())

		// the notarization arrives after instant finality already passed the nonce
		sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, []data.HeaderHandler{hdr1}, [][]byte{hash1})
		require.Equal(t, uint64(1), sfd.SettledCheckpointNonce())
		_, settledHash := sfd.GetHighestSettledBlockInfo()
		require.Equal(t, hash1, settledHash)
		require.Equal(t, uint64(2), sfd.FinalCheckpointNonce())

		sfd.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, []data.HeaderHandler{cleanHdr2}, [][]byte{hash2})
		require.Equal(t, uint64(2), sfd.SettledCheckpointNonce())
	})

	t.Run("andromeda settled watermark mirrors instant finality", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(andromedaOnlyHandler)

		_ = sfd.AddHeader(hdr1, hash1, process.BHProcessed, nil, nil)
		_ = sfd.AddHeader(cleanHdr2, hash2, process.BHProcessed, nil, nil)
		require.Equal(t, uint64(2), sfd.FinalCheckpointNonce())
		require.Equal(t, uint64(2), sfd.SettledCheckpointNonce())
	})

	t.Run("dead meta authority is removed and no longer protects the settled watermark", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForV3Settlement(supernovaHandler)
		headerHash := []byte("v3Hash")
		header := &block.HeaderV3{ShardID: 0, Epoch: 2, Nonce: 1, Round: 1}
		require.NoError(t, sfd.AddHeader(header, headerHash, process.BHProcessed, nil, nil))

		sfd.ReceivedSelfNotarizedFromCrossHeaders(
			core.MetachainShardId,
			[]data.HeaderHandler{header},
			[][]byte{headerHash},
		)
		require.Equal(t, uint64(1), sfd.SettledCheckpointNonce())
		require.Equal(t, headerHash, sfd.GetNotarizedHeaderHash(1))

		sfd.InvalidatedSelfNotarizedFromCrossHeaders(
			core.MetachainShardId,
			[]data.HeaderHandler{header},
			[][]byte{headerHash},
		)

		require.Equal(t, uint64(0), sfd.SettledCheckpointNonce())
		require.Nil(t, sfd.GetNotarizedHeaderHash(1))
		require.True(t, sfd.ReconcileFinalCheckpointBelow(1))
	})

	t.Run("invalidating consecutive authorities restores the last live settlement", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForV3Settlement(supernovaHandler)
		headers := []data.HeaderHandler{
			&block.HeaderV3{ShardID: 0, Epoch: 2, Nonce: 1, Round: 1},
			&block.HeaderV3{ShardID: 0, Epoch: 2, Nonce: 2, Round: 2, PrevHash: []byte("hash1")},
			&block.HeaderV3{ShardID: 0, Epoch: 2, Nonce: 3, Round: 3, PrevHash: []byte("hash2")},
		}
		hashes := [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")}
		for index, header := range headers {
			require.NoError(t, sfd.AddHeader(header, hashes[index], process.BHProcessed, nil, nil))
			sfd.ReceivedSelfNotarizedFromCrossHeaders(
				core.MetachainShardId,
				[]data.HeaderHandler{header},
				[][]byte{hashes[index]},
			)
		}
		require.Equal(t, uint64(3), sfd.SettledCheckpointNonce())

		sfd.InvalidatedSelfNotarizedFromCrossHeaders(
			core.MetachainShardId,
			headers[1:],
			hashes[1:],
		)

		require.Equal(t, uint64(1), sfd.SettledCheckpointNonce())
		require.Equal(t, uint64(3), sfd.FinalCheckpointNonce())
		_, settledHash := sfd.GetHighestSettledBlockInfo()
		require.Equal(t, hashes[0], settledHash)
		require.Nil(t, sfd.GetNotarizedHeaderHash(2))
		require.Nil(t, sfd.GetNotarizedHeaderHash(3))
	})

	t.Run("legacy notarization cannot be invalidated through the V3 path", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForV3Settlement(supernovaHandler)
		headerHash := []byte("legacyHash")
		header := &block.Header{ShardID: 0, Epoch: 2, Nonce: 1, Round: 1, PubKeysBitmap: []byte("X")}
		require.NoError(t, sfd.AddHeader(header, headerHash, process.BHProcessed, nil, nil))
		sfd.ReceivedSelfNotarizedFromCrossHeaders(
			core.MetachainShardId,
			[]data.HeaderHandler{header},
			[][]byte{headerHash},
		)

		sfd.InvalidatedSelfNotarizedFromCrossHeaders(
			core.MetachainShardId,
			[]data.HeaderHandler{header},
			[][]byte{headerHash},
		)

		require.Equal(t, uint64(1), sfd.SettledCheckpointNonce())
		require.Equal(t, headerHash, sfd.GetNotarizedHeaderHash(1))
	})

	t.Run("pre-activation V3 notarization cannot be invalidated", func(t *testing.T) {
		t.Parallel()

		sfd := createShardForkDetectorForFinality(supernovaHandler)
		headerHash := []byte("preActivationHash")
		header := &block.HeaderV3{ShardID: 0, Epoch: 2, Nonce: 1, Round: 1}
		require.NoError(t, sfd.AddHeader(header, headerHash, process.BHProcessed, nil, nil))
		sfd.ReceivedSelfNotarizedFromCrossHeaders(
			core.MetachainShardId,
			[]data.HeaderHandler{header},
			[][]byte{headerHash},
		)

		sfd.InvalidatedSelfNotarizedFromCrossHeaders(
			core.MetachainShardId,
			[]data.HeaderHandler{header},
			[][]byte{headerHash},
		)

		require.Equal(t, headerHash, sfd.GetNotarizedHeaderHash(1))
	})
}
