package interceptedBlocks

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createDefaultArgument() *ArgInterceptedBlockHeader {
	arg := &ArgInterceptedBlockHeader{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		MultiSigVerifier: mock.NewMultiSigner(),
		Hasher:           mock.HasherMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		ChronologyValidator: &mock.ChronologyValidatorStub{
			ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
				return nil
			},
		},
		HdrBuff: []byte("test buffer"),
	}

	return arg
}

func createDefaultHeaderHandler() *mock.HeaderHandlerStub {
	return &mock.HeaderHandlerStub{
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

//-------- checkArgument

func TestCheckArgument_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	err := checkArgument(nil)

	assert.Equal(t, process.ErrNilArguments, err)
}

func TestCheckArgument_NilHdrShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.HdrBuff = nil

	err := checkArgument(arg)

	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestCheckArgument_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.Marshalizer = nil

	err := checkArgument(arg)

	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestCheckArgument_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.Hasher = nil

	err := checkArgument(arg)

	assert.Equal(t, process.ErrNilHasher, err)
}

func TestCheckArgument_NilMultiSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.MultiSigVerifier = nil

	err := checkArgument(arg)

	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestCheckArgument_NilChronologyValidatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.ChronologyValidator = nil

	err := checkArgument(arg)

	assert.Equal(t, process.ErrNilChronologyValidator, err)
}

func TestCheckArgument_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.ShardCoordinator = nil

	err := checkArgument(arg)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestCheckArgument_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()

	err := checkArgument(arg)

	assert.Nil(t, err)
}

//-------- checkHeaderHandler

func TestCheckHeaderHandler_NilPubKeysBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetPubKeysBitmapCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr)

	assert.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestCheckHeaderHandler_NilPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetPrevHashCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr)

	assert.Equal(t, process.ErrNilPreviousBlockHash, err)
}

func TestCheckHeaderHandler_NilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetSignatureCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr)

	assert.Equal(t, process.ErrNilSignature, err)
}

func TestCheckHeaderHandler_NilRootHashErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetRootHashCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr)

	assert.Equal(t, process.ErrNilRootHash, err)
}

func TestCheckHeaderHandler_NilRandSeedErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetRandSeedCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr)

	assert.Equal(t, process.ErrNilRandSeed, err)
}

func TestCheckHeaderHandler_NilPrevRandSeedErr(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()
	hdr.GetPrevRandSeedCalled = func() []byte {
		return nil
	}

	err := checkHeaderHandler(hdr)

	assert.Equal(t, process.ErrNilPrevRandSeed, err)
}

func TestCheckHeaderHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createDefaultHeaderHandler()

	err := checkHeaderHandler(hdr)

	assert.Nil(t, err)
}

//------- checkMetaShardInfo

func TestCheckMetaShardInfo_WithNilOrEmptyShouldReturnNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	err1 := checkMetaShardInfo(nil, shardCoordinator)
	err2 := checkMetaShardInfo(make([]block.ShardData, 0), shardCoordinator)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func TestCheckMetaShardInfo_WrongShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	wrongShardId := uint32(2)
	sd := block.ShardData{
		ShardId:               wrongShardId,
		HeaderHash:            nil,
		ShardMiniBlockHeaders: nil,
		TxCount:               0,
	}

	err := checkMetaShardInfo([]block.ShardData{sd}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMetaShardInfo_WrongMiniblockSenderShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	wrongShardId := uint32(2)
	miniBlock := block.ShardMiniBlockHeader{
		Hash:            make([]byte, 0),
		ReceiverShardId: shardCoordinator.SelfId(),
		SenderShardId:   wrongShardId,
		TxCount:         0,
	}

	sd := block.ShardData{
		ShardId:               shardCoordinator.SelfId(),
		HeaderHash:            nil,
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{miniBlock},
		TxCount:               0,
	}

	err := checkMetaShardInfo([]block.ShardData{sd}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMetaShardInfo_WrongMiniblockReceiverShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	wrongShardId := uint32(2)
	miniBlock := block.ShardMiniBlockHeader{
		Hash:            make([]byte, 0),
		ReceiverShardId: wrongShardId,
		SenderShardId:   shardCoordinator.SelfId(),
		TxCount:         0,
	}

	sd := block.ShardData{
		ShardId:               shardCoordinator.SelfId(),
		HeaderHash:            nil,
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{miniBlock},
		TxCount:               0,
	}

	err := checkMetaShardInfo([]block.ShardData{sd}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMetaShardInfo_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	miniBlock := block.ShardMiniBlockHeader{
		Hash:            make([]byte, 0),
		ReceiverShardId: shardCoordinator.SelfId(),
		SenderShardId:   shardCoordinator.SelfId(),
		TxCount:         0,
	}

	sd := block.ShardData{
		ShardId:               shardCoordinator.SelfId(),
		HeaderHash:            nil,
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{miniBlock},
		TxCount:               0,
	}

	err := checkMetaShardInfo([]block.ShardData{sd}, shardCoordinator)

	assert.Nil(t, err)
}

//------- checkMiniblocks

func TestCheckMiniblocks_WithNilOrEmptyShouldReturnNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	err1 := checkMiniblocks(nil, shardCoordinator)
	err2 := checkMiniblocks(make([]block.MiniBlockHeader, 0), shardCoordinator)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func TestCheckMiniblocks_WrongMiniblockSenderShardIdShouldErr(t *testing.T) {
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

	err := checkMiniblocks([]block.MiniBlockHeader{miniblockHeader}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMiniblocks_WrongMiniblockReceiverShardIdShouldErr(t *testing.T) {
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

	err := checkMiniblocks([]block.MiniBlockHeader{miniblockHeader}, shardCoordinator)

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestCheckMiniblocks_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	miniblockHeader := block.MiniBlockHeader{
		Hash:            make([]byte, 0),
		SenderShardID:   shardCoordinator.SelfId(),
		ReceiverShardID: shardCoordinator.SelfId(),
		TxCount:         0,
		Type:            0,
	}

	err := checkMiniblocks([]block.MiniBlockHeader{miniblockHeader}, shardCoordinator)

	assert.Nil(t, err)
}
