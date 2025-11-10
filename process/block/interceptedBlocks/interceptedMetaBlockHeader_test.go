package interceptedBlocks_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/graceperiod"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func createDefaultMetaArgument() *interceptedBlocks.ArgInterceptedBlockHeader {
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	return createMetaArgumentWithShardCoordinator(shardCoordinator)
}

func createDefaultMetaV3Argument() *interceptedBlocks.ArgInterceptedBlockHeader {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	arg := createMetaV3ArgumentWithShardCoordinator(shardCoordinator)
	arg.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		},
	}
	return arg
}

func createMetaArgumentWithShardCoordinatorAndHeader(shardCoordinator sharding.Coordinator, hdr data.MetaHeaderHandler) *interceptedBlocks.ArgInterceptedBlockHeader {
	gracePeriod, _ := graceperiod.NewEpochChangeGracePeriod([]config.EpochChangeGracePeriodByEpoch{{EnableEpoch: 0, GracePeriodInRounds: 1}})
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator:        shardCoordinator,
		Hasher:                  testHasher,
		Marshalizer:             testMarshalizer,
		HeaderSigVerifier:       &consensus.HeaderSigVerifierMock{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		ValidityAttester:        &mock.ValidityAttesterStub{},
		EpochStartTrigger: &mock.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return hdrEpoch
			},
		},
		EnableEpochsHandler:           &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EpochChangeGracePeriodHandler: gracePeriod,
	}
	if hdr.IsHeaderV3() {
		arg.HdrBuff, _ = testMarshalizer.Marshal(hdr.(*dataBlock.MetaBlockV3))
	} else {
		arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)
	}

	return arg
}

func createMetaArgumentWithShardCoordinator(shardCoordinator sharding.Coordinator) *interceptedBlocks.ArgInterceptedBlockHeader {
	hdr := createMockMetaHeader()
	return createMetaArgumentWithShardCoordinatorAndHeader(shardCoordinator, hdr)
}

func createMetaV3ArgumentWithShardCoordinator(shardCoordinator sharding.Coordinator) *interceptedBlocks.ArgInterceptedBlockHeader {
	hdr := createMockMetaHeaderV3()
	return createMetaArgumentWithShardCoordinatorAndHeader(shardCoordinator, hdr)
}

func createMockMetaHeader() *dataBlock.MetaBlock {
	return &dataBlock.MetaBlock{
		Nonce:                  hdrNonce,
		PrevHash:               []byte("prev hash"),
		PrevRandSeed:           []byte("prev rand seed"),
		RandSeed:               []byte("rand seed"),
		PubKeysBitmap:          []byte{1},
		TimeStamp:              0,
		Round:                  hdrRound,
		Epoch:                  hdrEpoch,
		Signature:              []byte("signature"),
		RootHash:               []byte("root hash"),
		TxCount:                0,
		ShardInfo:              nil,
		ChainID:                []byte("chain ID"),
		SoftwareVersion:        []byte("software version"),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		ValidatorStatsRootHash: []byte("validator stats root hash"),
	}
}

func createMockMetaHeaderV3() *dataBlock.MetaBlockV3 {
	return &dataBlock.MetaBlockV3{
		Nonce:           42,
		Epoch:           2,
		Round:           15,
		TimestampMs:     123456789,
		PrevHash:        []byte("prev_hash"),
		PrevRandSeed:    []byte("prev_seed"),
		RandSeed:        []byte("new_seed"),
		ChainID:         []byte("chain-id"),
		SoftwareVersion: []byte("v1.0.0"),
		LeaderSignature: []byte("leader_signature"),

		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{Hash: []byte("meta-to-s0"), SenderShardID: core.MetachainShardId, ReceiverShardID: 0},
			{Hash: []byte("meta-to-s1"), SenderShardID: core.MetachainShardId, ReceiverShardID: 1},
		},

		ShardInfo: []dataBlock.ShardData{
			{
				ShardID:    0,
				Round:      10,
				Nonce:      41,
				Epoch:      1,
				HeaderHash: []byte("shard0-hash"),
				ShardMiniBlockHeaders: []dataBlock.MiniBlockHeader{
					{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("s0-to-s1")},
				},
			},
			{
				ShardID:    1,
				Round:      11,
				Nonce:      40,
				Epoch:      1,
				HeaderHash: []byte("shard1-hash"),
				ShardMiniBlockHeaders: []dataBlock.MiniBlockHeader{
					{SenderShardID: 1, ReceiverShardID: 0, Hash: []byte("s1-to-s0")},
				},
			},
		},
		ShardInfoProposal: []dataBlock.ShardDataProposal{
			{ShardID: 0, HeaderHash: []byte("shard-0-hash"), Nonce: 41, Round: 10, Epoch: 1},
			{ShardID: 1, HeaderHash: []byte("shard-1-hash"), Nonce: 40, Round: 11, Epoch: 1},
		},
		ExecutionResults: []*dataBlock.MetaExecutionResult{
			{
				ExecutionResult: &dataBlock.BaseMetaExecutionResult{
					BaseExecutionResult: &dataBlock.BaseExecutionResult{
						HeaderHash:  []byte("hdr-hash-10"),
						HeaderNonce: 39,
						HeaderRound: 10,
						HeaderEpoch: 1,
						RootHash:    []byte("root-hash-10"),
					},
				},
			},
			{
				ExecutionResult: &dataBlock.BaseMetaExecutionResult{
					BaseExecutionResult: &dataBlock.BaseExecutionResult{
						HeaderHash:  []byte("hdr-hash-11"),
						HeaderNonce: 40,
						HeaderRound: 11,
						HeaderEpoch: 1,
						RootHash:    []byte("root-hash-11"),
					},
				},
			},
			{
				ExecutionResult: &dataBlock.BaseMetaExecutionResult{
					BaseExecutionResult: &dataBlock.BaseExecutionResult{
						HeaderHash:  []byte("hdr-hash-last"),
						HeaderNonce: 41,
						HeaderRound: 12,
						HeaderEpoch: 2,
						RootHash:    []byte("root-hash-last"),
					},
				},
			},
		},

		LastExecutionResult: &dataBlock.MetaExecutionResultInfo{
			NotarizedInRound: 14,
			ExecutionResult: &dataBlock.BaseMetaExecutionResult{
				BaseExecutionResult: &dataBlock.BaseExecutionResult{
					HeaderHash:  []byte("hdr-hash-last"),
					HeaderNonce: 41,
					HeaderRound: 12,
					HeaderEpoch: 2,
					RootHash:    []byte("root-hash-last"),
				},
			},
		},
	}
}

// ------- TestNewInterceptedHeader

func TestNewInterceptedMetaHeader_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	inHdr, err := interceptedBlocks.NewInterceptedMetaHeader(nil)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedMetaHeader_MarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	arg.HdrBuff = []byte("invalid buffer")

	inHdr, err := interceptedBlocks.NewInterceptedMetaHeader(arg)

	assert.Nil(t, inHdr)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestNewInterceptedMetaHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()

	inHdr, err := interceptedBlocks.NewInterceptedMetaHeader(arg)

	assert.False(t, check.IfNil(inHdr))
	assert.Nil(t, err)
}

// ------- CheckValidity

func TestInterceptedMetaHeader_CheckValidityNilPubKeyBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockMetaHeader()
	hdr.PubKeysBitmap = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultMetaArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestInterceptedMetaHeader_ErrorInMiniBlockShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockMetaHeader()
	badShardId := uint32(2)
	hdr.ShardInfo = []dataBlock.ShardData{
		{
			ShardID:               badShardId,
			HeaderHash:            nil,
			ShardMiniBlockHeaders: nil,
			TxCount:               0,
		},
	}
	buff, _ := testMarshalizer.Marshal(hdr)

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)

	arg := createMetaArgumentWithShardCoordinator(shardCoordinator)
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedMetaHeader_CheckValidityShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()

	assert.Nil(t, err)
}

func TestInterceptedMetaHeader_CheckAgainstRoundHandlerAttesterFailsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	expectedErr := errors.New("expected error")
	arg.ValidityAttester = &mock.ValidityAttesterStub{
		CheckBlockAgainstRoundHandlerCalled: func(headerHandler data.HeaderHandler) error {
			return expectedErr
		},
	}
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, expectedErr, err)
}

func TestInterceptedMetaHeader_CheckAgainstFinalHeaderAttesterFailsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	expectedErr := errors.New("expected error")
	arg.ValidityAttester = &mock.ValidityAttesterStub{
		CheckBlockAgainstFinalCalled: func(headerHandler data.HeaderHandler) error {
			return expectedErr
		},
	}
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, expectedErr, err)
}

// ------- getters

func TestInterceptedMetaHeader_Getters(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	hash := testHasher.Compute(string(arg.HdrBuff))

	assert.Equal(t, hash, inHdr.Hash())
	assert.True(t, inHdr.IsForCurrentShard())
	require.False(t, inHdr.ShouldAllowDuplicates())
}

func TestInterceptedMetaHeader_CheckValidityLeaderSignatureNotCorrectShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockMetaHeader()
	expectedErr := errors.New("expected err")
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultMetaArgument()
	arg.HeaderSigVerifier = &consensus.HeaderSigVerifierMock{
		VerifyRandSeedAndLeaderSignatureCalled: func(header data.HeaderHandler) error {
			return expectedErr
		},
	}
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()
	assert.Equal(t, expectedErr, err)
}
func TestInterceptedMetaHeader_CheckValidityErrorInMiniBlockShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockMetaHeader()
	badShardId := uint32(2)
	hdr.MiniBlockHeaders = []dataBlock.MiniBlockHeader{
		{
			Hash:            make([]byte, 0),
			SenderShardID:   badShardId,
			ReceiverShardID: 0,
			TxCount:         0,
			Type:            0,
		},
	}

	arg := createDefaultMetaArgument()
	marshaller := arg.Marshalizer
	buff, _ := marshaller.Marshal(hdr)
	arg.HdrBuff = buff
	inHdr, err := interceptedBlocks.NewInterceptedMetaHeader(arg)
	require.Nil(t, err)
	require.NotNil(t, inHdr)

	err = inHdr.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedMetaHeader_CheckValidityLeaderSignatureOkShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createMockMetaHeader()
	expectedSignature := []byte("ran")
	hdr.LeaderSignature = expectedSignature
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultMetaArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()
	assert.Nil(t, err)
}

func TestInterceptedMetaHeader_CheckValidityExecutionResultMiniblockErrorInMetaHeaderV3ShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaV3Argument()
	inHdr, err := interceptedBlocks.NewInterceptedMetaHeader(arg)
	assert.Nil(t, err)
	assert.NotNil(t, inHdr)

	assert.True(t, inHdr.HeaderHandler().IsHeaderV3())
	badShardId := uint32(28)
	mbs := []dataBlock.MiniBlockHeader{
		{
			Hash:            make([]byte, 0),
			SenderShardID:   badShardId,
			ReceiverShardID: 0,
			TxCount:         0,
			Type:            0,
		},
	}
	mbHandlers := make([]data.MiniBlockHeaderHandler, len(mbs))
	for i := range mbs {
		tmp := mbs[i]
		mbHandlers[i] = &tmp
	}
	_ = inHdr.HeaderHandler().(*dataBlock.MetaBlockV3).ExecutionResults[1].SetMiniBlockHeadersHandlers(mbHandlers)
	err = inHdr.CheckValidity()

	assert.Error(t, err)
	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedMetaHeader_CheckValidityShouldWorkV3(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaV3Argument()
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)
	assert.True(t, inHdr.HeaderHandler().IsHeaderV3())
	err := inHdr.CheckValidity()
	assert.Nil(t, err)
}

func TestInterceptedMetaHeader_isMetaHeaderEpochOutOfRange(t *testing.T) {
	epochStartTrigger := &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 10
		},
	}
	t.Run("old epoch header accepted", func(t *testing.T) {
		arg := createDefaultMetaArgument()
		arg.EpochStartTrigger = epochStartTrigger
		hdr := createMockMetaHeader()
		hdr.Epoch = 8
		arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

		inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)
		require.False(t, inHdr.IsMetaHeaderOutOfRange())
	})

	t.Run("current epoch header accepted", func(t *testing.T) {
		arg := createDefaultMetaArgument()
		arg.EpochStartTrigger = epochStartTrigger
		hdr := createMockMetaHeader()
		hdr.Epoch = 10
		arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

		inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)
		require.False(t, inHdr.IsMetaHeaderOutOfRange())
	})

	t.Run("next epoch header accepted", func(t *testing.T) {
		arg := createDefaultMetaArgument()
		arg.EpochStartTrigger = epochStartTrigger
		hdr := createMockMetaHeader()
		hdr.Epoch = 11
		arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

		inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)
		require.False(t, inHdr.IsMetaHeaderOutOfRange())
	})

	t.Run("larger epoch difference header rejected", func(t *testing.T) {
		arg := createDefaultMetaArgument()
		arg.EpochStartTrigger = epochStartTrigger
		hdr := createMockMetaHeader()
		hdr.Epoch = 12
		arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

		inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)
		require.True(t, inHdr.IsMetaHeaderOutOfRange())
	})
}

// ------- IsInterfaceNil

func TestInterceptedMetaHeader_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var inHdr *interceptedBlocks.InterceptedMetaHeader

	assert.True(t, check.IfNil(inHdr))
}
