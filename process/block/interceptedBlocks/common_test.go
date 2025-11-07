package interceptedBlocks

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/common/graceperiod"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
)

func createDefaultBlockHeaderArgument() *ArgInterceptedBlockHeader {
	gracePeriod, _ := graceperiod.NewEpochChangeGracePeriod([]config.EpochChangeGracePeriodByEpoch{{EnableEpoch: 0, GracePeriodInRounds: 1}})
	arg := &ArgInterceptedBlockHeader{
		ShardCoordinator:              mock.NewOneShardCoordinatorMock(),
		Hasher:                        &hashingMocks.HasherMock{},
		Marshalizer:                   &mock.MarshalizerMock{},
		HdrBuff:                       []byte("test buffer"),
		HeaderSigVerifier:             &consensus.HeaderSigVerifierMock{},
		HeaderIntegrityVerifier:       &mock.HeaderIntegrityVerifierStub{},
		ValidityAttester:              &mock.ValidityAttesterStub{},
		EpochStartTrigger:             &mock.EpochStartTriggerStub{},
		EnableEpochsHandler:           &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EpochChangeGracePeriodHandler: gracePeriod,
	}

	return arg
}

func createDefaultMiniblockArgument() *ArgInterceptedMiniblock {
	arg := &ArgInterceptedMiniblock{
		Hasher:           &hashingMocks.HasherMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		MiniblockBuff:    []byte("test buffer"),
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
	}

	return arg
}

func createDefaultHeaderHandler() *testscommon.HeaderHandlerStub {
	return &testscommon.HeaderHandlerStub{
		GetPubKeysBitmapCalled: func() []byte {
			return []byte("pub keys bitmap")
		},
		GetSignatureCalled: func() []byte {
			return []byte("signature")
		},
		GetRootHashCalled: func() []byte {
			return []byte("root hash")
		},
		GetRandSeedCalled: func() []byte {
			return []byte("rand seed")
		},
		GetPrevRandSeedCalled: func() []byte {
			return []byte("prev rand seed")
		},
		GetPrevHashCalled: func() []byte {
			return []byte("prev hash")
		},
	}
}

func createValidHeaderV3ToTest() *block.HeaderV3 {
	return &block.HeaderV3{
		Nonce:           101,
		Round:           1200,
		Epoch:           1,
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		LeaderSignature: []byte("leader signature"),
		SoftwareVersion: []byte("v1.0.0"),

		ExecutionResults: []*block.ExecutionResult{
			&block.ExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("header hash"),
					HeaderNonce: 98,
					HeaderRound: 1001,
					HeaderEpoch: 0,
					RootHash:    []byte("root hash"),
				},
			},
			&block.ExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("header hash"),
					HeaderNonce: 99,
					HeaderRound: 1020,
					HeaderEpoch: 0,
					RootHash:    []byte("root hash"),
				},
			},
			&block.ExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("header hash"),
					HeaderNonce: 100,
					HeaderRound: 1100,
					HeaderEpoch: 1,
					RootHash:    []byte("root hash"),
				},
			},
		},
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("header hash"),
				HeaderNonce: 100,
				HeaderRound: 1100,
				HeaderEpoch: 1,
				RootHash:    []byte("root hash"),
			},
			NotarizedInRound: 1199,
		},
	}
}

func createValidMetaHeaderV3ToTest() *block.MetaBlockV3 {
	return &block.MetaBlockV3{
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

		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte("meta-to-s0"), SenderShardID: core.MetachainShardId, ReceiverShardID: 0},
			{Hash: []byte("meta-to-s1"), SenderShardID: core.MetachainShardId, ReceiverShardID: 1},
		},

		ShardInfo: []block.ShardData{
			{
				ShardID:    0,
				Round:      10,
				Nonce:      41,
				Epoch:      1,
				HeaderHash: []byte("shard0-hash"),
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("s0-to-s1")},
				},
			},
			{
				ShardID:    1,
				Round:      11,
				Nonce:      40,
				Epoch:      1,
				HeaderHash: []byte("shard1-hash"),
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{SenderShardID: 1, ReceiverShardID: 0, Hash: []byte("s1-to-s0")},
				},
			},
		},
		ShardInfoProposal: []block.ShardDataProposal{
			{ShardID: 0, HeaderHash: []byte("shard-0-hash"), Nonce: 41, Round: 10, Epoch: 1},
			{ShardID: 1, HeaderHash: []byte("shard-1-hash"), Nonce: 40, Round: 11, Epoch: 1},
		},
		ExecutionResults: []*block.MetaExecutionResult{
			{
				ExecutionResult: &block.BaseMetaExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash:  []byte("hdr-hash-10"),
						HeaderNonce: 39,
						HeaderRound: 10,
						HeaderEpoch: 1,
						RootHash:    []byte("root-hash-10"),
					},
				},
			},
			{
				ExecutionResult: &block.BaseMetaExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash:  []byte("hdr-hash-11"),
						HeaderNonce: 40,
						HeaderRound: 11,
						HeaderEpoch: 1,
						RootHash:    []byte("root-hash-11"),
					},
				},
			},
			{
				ExecutionResult: &block.BaseMetaExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash:  []byte("hdr-hash-last"),
						HeaderNonce: 41,
						HeaderRound: 12,
						HeaderEpoch: 2,
						RootHash:    []byte("root-hash-last"),
					},
				},
			},
		},

		LastExecutionResult: &block.MetaExecutionResultInfo{
			NotarizedInRound: 14,
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
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

// -------- checkBlockHeaderArgument

func TestCheckBlockHeaderArgument_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	err := checkBlockHeaderArgument(nil)

	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestCheckBlockHeaderArgument_NilHdrShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.HdrBuff = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestCheckBlockHeaderArgument_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.Marshalizer = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestCheckBlockHeaderArgument_NilHeaderSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.HeaderSigVerifier = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilHeaderSigVerifier, err)
}

func TestCheckBlockHeaderArgument_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.Hasher = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilHasher, err)
}

func TestCheckBlockHeaderArgument_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.ShardCoordinator = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestCheckBlockHeaderArgument_NilHeaderIntegrityVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.HeaderIntegrityVerifier = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilHeaderIntegrityVerifier, err)
}

func TestCheckBlockHeaderArgument_NilEpochStartTriggerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.EpochStartTrigger = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilEpochStartTrigger, err)
}

func TestCheckBlockHeaderArgument_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.ValidityAttester = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilValidityAttester, err)
}

func TestCheckBlockHeaderArgument_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()
	arg.EnableEpochsHandler = nil

	err := checkBlockHeaderArgument(arg)

	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestCheckBlockHeaderArgument_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultBlockHeaderArgument()

	err := checkBlockHeaderArgument(arg)

	assert.Nil(t, err)
}

// -------- checkMiniblockArgument

func TestCheckMiniblockArgument_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	err := checkMiniblockArgument(nil)

	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestCheckMiniblockArgument_NilHdrShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = nil

	err := checkMiniblockArgument(arg)

	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestCheckMiniblockArgument_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	arg.Marshalizer = nil

	err := checkMiniblockArgument(arg)

	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestCheckMiniblockArgument_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	arg.Hasher = nil

	err := checkMiniblockArgument(arg)

	assert.Equal(t, process.ErrNilHasher, err)
}

func TestCheckMiniblockArgument_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	arg.ShardCoordinator = nil

	err := checkMiniblockArgument(arg)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestCheckMiniblockArgument_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()

	err := checkMiniblockArgument(arg)

	assert.Nil(t, err)
}

// -------- checkHeaderHandler

func TestCheckHeaderHandler_NilPubKeysBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetPubKeysBitmapCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestCheckHeaderHandler_NilPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetPrevHashCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Equal(t, process.ErrNilPreviousBlockHash, err)
}

func TestCheckHeaderHandler_NilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetSignatureCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Equal(t, process.ErrNilSignature, err)
}

func TestCheckHeaderHandler_NilRootHashErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetRootHashCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Equal(t, process.ErrNilRootHash, err)
}

func TestCheckHeaderHandler_NilRandSeedErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetRandSeedCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Equal(t, process.ErrNilRandSeed, err)
}

func TestCheckHeaderHandler_NilPrevRandSeedErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetPrevRandSeedCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Equal(t, process.ErrNilPrevRandSeed, err)
}

func TestCheckHeaderHandler_CheckFieldsForNilErrors(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.CheckFieldsForNilCalled = func() error {
		return expectedErr
	}

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Equal(t, expectedErr, err)
}

func TestCheckHeaderHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()

	err := checkHeaderHandler(hdr, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())

	assert.Nil(t, err)
}

func TestCheckHeaderHandler_HeaderV3_CheckFieldsIntegrityErrors(t *testing.T) {
	t.Parallel()

	enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
		return true
	}
	t.Run("nil header", func(t *testing.T) {
		t.Parallel()
		var hv3 *block.HeaderV3
		err := checkHeaderHandler(hv3, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("not nil receipt hash field", func(t *testing.T) {
		t.Parallel()
		hv3 := createValidHeaderV3ToTest()
		hv3.ReceiptsHash = []byte("not nil receipts hash")
		err := checkHeaderHandler(hv3, enableEpochsHandler)
		assert.True(t, hv3.IsHeaderV3())
		assert.Error(t, err)
		assert.ErrorIs(t, err, data.ErrNotNilValue)
	})
	t.Run("not nil reserved field", func(t *testing.T) {
		t.Parallel()
		hv3 := createValidHeaderV3ToTest()
		hv3.Reserved = []byte("not nil reserved")
		err := checkHeaderHandler(hv3, enableEpochsHandler)
		assert.Error(t, err)
		assert.ErrorIs(t, err, data.ErrNotNilValue)
	})
	t.Run("nil last execution result", func(t *testing.T) {
		t.Parallel()
		hv3 := &block.HeaderV3{
			Nonce: 2,
			Round: 2,
			Epoch: 2,
		}
		err := checkHeaderHandler(hv3, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("invalid execution results", func(t *testing.T) {
		t.Parallel()
		hv3 := createValidHeaderV3ToTest()
		hv3.ExecutionResults[0].BaseExecutionResult = nil
		err := checkHeaderHandler(hv3, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("invalid last execution result", func(t *testing.T) {
		t.Parallel()
		hv3 := createValidHeaderV3ToTest()
		hv3.LastExecutionResult.ExecutionResult = nil
		err := checkHeaderHandler(hv3, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("invalid round in last execution result", func(t *testing.T) {
		t.Parallel()
		hv3 := createValidHeaderV3ToTest()
		hv3.LastExecutionResult.NotarizedInRound = 1300
		err := checkHeaderHandler(hv3, enableEpochsHandler)
		assert.Error(t, err)
	})
}
func TestCheckHeaderHandler_HeaderV3_CheckFieldsIntegrityWorks(t *testing.T) {
	t.Parallel()
	hdr := createValidHeaderV3ToTest()
	enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
		return true
	}
	err := checkHeaderHandler(hdr, enableEpochsHandler)
	assert.Nil(t, err)
}

func TestCheckHeaderHandler_MetaHeaderV3_CheckFieldsIntegrityErrors(t *testing.T) {
	t.Parallel()

	enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
		return true
	}
	t.Run("nil header", func(t *testing.T) {
		t.Parallel()
		var meta *block.MetaBlockV3
		err := checkHeaderHandler(meta, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("not nil reserved field", func(t *testing.T) {
		t.Parallel()
		meta := createValidMetaHeaderV3ToTest()
		meta.Reserved = []byte("not nil reserved")
		err := checkHeaderHandler(meta, enableEpochsHandler)
		assert.Error(t, err)
		assert.ErrorIs(t, err, data.ErrNotNilValue)
	})
	t.Run("nil last execution result", func(t *testing.T) {
		t.Parallel()
		meta := &block.MetaBlockV3{
			Nonce: 2,
			Round: 2,
			Epoch: 2,
		}
		err := checkHeaderHandler(meta, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("invalid execution results", func(t *testing.T) {
		t.Parallel()
		meta := createValidMetaHeaderV3ToTest()
		meta.ExecutionResults[0].ExecutionResult = nil
		err := checkHeaderHandler(meta, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("invalid last execution result", func(t *testing.T) {
		t.Parallel()
		meta := createValidMetaHeaderV3ToTest()
		meta.LastExecutionResult.ExecutionResult = nil
		err := checkHeaderHandler(meta, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("invalid round in last execution result", func(t *testing.T) {
		t.Parallel()
		meta := createValidMetaHeaderV3ToTest()
		meta.LastExecutionResult.NotarizedInRound = 1300
		err := checkHeaderHandler(meta, enableEpochsHandler)
		assert.Error(t, err)
	})
	t.Run("invalid shard info", func(t *testing.T) {
		t.Parallel()
		meta := createValidMetaHeaderV3ToTest()
		meta.ShardInfoProposal = make([]block.ShardDataProposal, 0)
		err := checkHeaderHandler(meta, enableEpochsHandler)
		assert.Error(t, err)
	})
}

func TestCheckHeaderHandler_MetaHeaderV3_CheckFieldsIntegrityWorks(t *testing.T) {
	t.Parallel()
	hdr := createValidMetaHeaderV3ToTest()
	enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
		return true
	}
	err := checkHeaderHandler(hdr, enableEpochsHandler)
	assert.Nil(t, err)
}

// ------- checkMetaShardInfo

func TestCheckMetaShardInfo_WithNilOrEmptyShouldReturnNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	err1 := checkMetaShardInfo(nil, shardCoordinator)
	err2 := checkMetaShardInfo(make([]data.ShardDataHandler, 0), shardCoordinator)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func TestCheckMetaShardInfo_ShouldNotCheckShardInfoForShards(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(1)

	sd := block.ShardData{}

	err := checkMetaShardInfo([]data.ShardDataHandler{&sd}, shardCoordinator)
	assert.Nil(t, err)
}

func TestCheckMetaShardInfo_WrongShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)
	wrongShardId := uint32(2)
	sd := block.ShardData{
		ShardID:               wrongShardId,
		HeaderHash:            nil,
		ShardMiniBlockHeaders: nil,
		TxCount:               0,
	}

	err := checkMetaShardInfo([]data.ShardDataHandler{&sd}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMetaShardInfo_WrongMiniblockSenderShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)
	wrongShardId := uint32(2)
	miniBlock := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		ReceiverShardID: shardCoordinator.SelfId(),
		SenderShardID:   wrongShardId,
		TxCount:         0,
	}

	sd := block.ShardData{
		ShardID:               shardCoordinator.SelfId(),
		HeaderHash:            nil,
		ShardMiniBlockHeaders: []block.MiniBlockHeader{miniBlock},
		TxCount:               0,
	}

	err := checkMetaShardInfo([]data.ShardDataHandler{&sd}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMetaShardInfo_WrongMiniblockReceiverShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)
	wrongShardId := uint32(2)
	miniBlock := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		ReceiverShardID: wrongShardId,
		SenderShardID:   shardCoordinator.SelfId(),
		TxCount:         0,
	}

	sd := block.ShardData{
		ShardID:               shardCoordinator.SelfId(),
		HeaderHash:            nil,
		ShardMiniBlockHeaders: []block.MiniBlockHeader{miniBlock},
		TxCount:               0,
	}

	err := checkMetaShardInfo([]data.ShardDataHandler{&sd}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMetaShardInfo_ReservedPopulatedShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)

	miniBlock := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		ReceiverShardID: shardCoordinator.SelfId(),
		SenderShardID:   shardCoordinator.SelfId(),
		TxCount:         0,
		Reserved:        []byte("rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr"),
	}

	sd := block.ShardData{
		ShardID:               shardCoordinator.SelfId(),
		HeaderHash:            nil,
		ShardMiniBlockHeaders: []block.MiniBlockHeader{miniBlock},
		TxCount:               0,
	}

	err := checkMetaShardInfo([]data.ShardDataHandler{&sd}, shardCoordinator)

	assert.Equal(t, process.ErrReservedFieldInvalid, err)
}

func TestCheckMetaShardInfo_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)
	miniBlock := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		ReceiverShardID: shardCoordinator.SelfId(),
		SenderShardID:   shardCoordinator.SelfId(),
		TxCount:         0,
	}

	sd := block.ShardData{
		ShardID:               shardCoordinator.SelfId(),
		HeaderHash:            nil,
		ShardMiniBlockHeaders: []block.MiniBlockHeader{miniBlock},
		TxCount:               0,
	}

	err := checkMetaShardInfo([]data.ShardDataHandler{&sd}, shardCoordinator)
	assert.Nil(t, err)

	miniBlock.Reserved = []byte("r")
	sd.ShardMiniBlockHeaders = []block.MiniBlockHeader{miniBlock}
	err = checkMetaShardInfo([]data.ShardDataHandler{&sd}, shardCoordinator)
	assert.Nil(t, err)
}

func TestCheckMetaShardInfo_WithMultipleShardData(t *testing.T) {
	t.Parallel()

	t.Run("should return invalid shard id error, with multiple shard data", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewOneShardCoordinatorMock()
		_ = shardCoordinator.SetSelfId(core.MetachainShardId)
		wrongShardId := uint32(2)
		miniBlock1 := block.MiniBlockHeader{
			Hash:            make([]byte, 0),
			ReceiverShardID: wrongShardId,
			SenderShardID:   shardCoordinator.SelfId(),
			TxCount:         0,
		}

		miniBlock2 := block.MiniBlockHeader{
			Hash:            make([]byte, 0),
			ReceiverShardID: shardCoordinator.SelfId(),
			SenderShardID:   shardCoordinator.SelfId(),
			TxCount:         0,
		}

		sd1 := &block.ShardData{
			ShardID:    shardCoordinator.SelfId(),
			HeaderHash: nil,
			ShardMiniBlockHeaders: []block.MiniBlockHeader{
				miniBlock2,
			},
			TxCount: 0,
		}

		sd2 := &block.ShardData{
			ShardID:    shardCoordinator.SelfId(),
			HeaderHash: nil,
			ShardMiniBlockHeaders: []block.MiniBlockHeader{
				miniBlock1,
			},
			TxCount: 0,
		}

		err := checkMetaShardInfo(
			[]data.ShardDataHandler{sd1, sd2},
			shardCoordinator,
		)

		assert.Equal(t, process.ErrInvalidShardId, err)
	})
}

// ------- checkMiniBlocksHeaders

func TestCheckMiniBlocksHeaders_WithNilOrEmptyShouldReturnNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	err1 := checkMiniBlocksHeaders(nil, shardCoordinator)
	err2 := checkMiniBlocksHeaders(make([]data.MiniBlockHeaderHandler, 0), shardCoordinator)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func TestCheckMiniBlocksHeaders_WrongMiniblockSenderShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	wrongShardId := uint32(2)
	miniblockHeader := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		SenderShardID:   wrongShardId,
		ReceiverShardID: shardCoordinator.SelfId(),
		TxCount:         0,
		Type:            0,
	}

	err := checkMiniBlocksHeaders([]data.MiniBlockHeaderHandler{&miniblockHeader}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMiniBlocksHeaders_WrongMiniblockReceiverShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	wrongShardId := uint32(2)
	miniblockHeader := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		SenderShardID:   shardCoordinator.SelfId(),
		ReceiverShardID: wrongShardId,
		TxCount:         0,
		Type:            0,
	}

	err := checkMiniBlocksHeaders([]data.MiniBlockHeaderHandler{&miniblockHeader}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMiniBlocksHeaders_ReservedPopulatedShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	miniblockHeader := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		SenderShardID:   shardCoordinator.SelfId(),
		ReceiverShardID: shardCoordinator.SelfId(),
		TxCount:         0,
		Type:            0,
		Reserved:        []byte("rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr"),
	}

	err := checkMiniBlocksHeaders([]data.MiniBlockHeaderHandler{&miniblockHeader}, shardCoordinator)

	assert.Equal(t, process.ErrReservedFieldInvalid, err)
}

func TestCheckMiniBlocksHeaders_ReservedPopulatedCorrectly(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	miniblockHeader := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		SenderShardID:   shardCoordinator.SelfId(),
		ReceiverShardID: shardCoordinator.SelfId(),
		TxCount:         0,
		Type:            0,
		Reserved:        []byte("r"),
	}

	err := checkMiniBlocksHeaders([]data.MiniBlockHeaderHandler{&miniblockHeader}, shardCoordinator)

	assert.Nil(t, err)
}

func TestCheckMiniBlocksHeaders_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	miniblockHeader := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		SenderShardID:   shardCoordinator.SelfId(),
		ReceiverShardID: shardCoordinator.SelfId(),
		TxCount:         0,
		Type:            0,
	}

	err := checkMiniBlocksHeaders([]data.MiniBlockHeaderHandler{&miniblockHeader}, shardCoordinator)

	assert.Nil(t, err)
}
