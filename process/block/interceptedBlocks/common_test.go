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
