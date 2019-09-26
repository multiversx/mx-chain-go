package interceptedBlocks_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createDefaultMetaArgument() *interceptedBlocks.ArgInterceptedBlockHeader {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		MultiSigVerifier: mock.NewMultiSigner(),
		Hasher:           testHasher,
		Marshalizer:      testMarshalizer,
		ChronologyValidator: &mock.ChronologyValidatorStub{
			ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
				return nil
			},
		},
	}

	hdr := createMockMetaHeader()
	arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

	return arg
}

func createMockMetaHeader() *dataBlock.MetaBlock {
	return &dataBlock.MetaBlock{
		Nonce:         hdrNonce,
		PrevHash:      []byte("prev hash"),
		PrevRandSeed:  []byte("prev rand seed"),
		RandSeed:      []byte("rand seed"),
		PubKeysBitmap: []byte("bitmap"),
		TimeStamp:     0,
		Round:         hdrRound,
		Epoch:         hdrEpoch,
		Signature:     []byte("signature"),
		RootHash:      []byte("root hash"),
		TxCount:       0,
		PeerInfo:      nil,
		ShardInfo:     nil,
	}
}

//------- TestNewInterceptedHeader

func TestNewInterceptedMetaHeader_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	inHdr, err := interceptedBlocks.NewInterceptedMetaHeader(nil)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilArguments, err)
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

//------- CheckValidity

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
			ShardId:               badShardId,
			HeaderHash:            nil,
			ShardMiniBlockHeaders: nil,
			TxCount:               0,
		},
	}
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultShardArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestNewInterceptedMetaHeader_CheckValidityShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	err := inHdr.CheckValidity()

	assert.Nil(t, err)
}

//------- getters

func TestInterceptedMetaHeader_Getters(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	hash := testHasher.Compute(string(arg.HdrBuff))

	assert.Equal(t, hash, inHdr.Hash())
	assert.True(t, inHdr.IsForCurrentShard())
}

//------- IsInterfaceNil

func TestInterceptedMetaHeader_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var inHdr *interceptedBlocks.InterceptedMetaHeader

	assert.True(t, check.IfNil(inHdr))
}

//
//func TestInterceptedMetaHeader_IntegrityNilRootHashShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = nil
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//
//	assert.Equal(t, process.ErrNilRootHash, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityNilPrevRandHashShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = nil
//	hdr.RandSeed = make([]byte, 0)
//
//	assert.Equal(t, process.ErrNilPrevRandSeed, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityNilRandHashShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = nil
//
//	assert.Equal(t, process.ErrNilRandSeed, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityNilInvalidShardIdOnShardedDataShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//	hdr.ShardInfo = []block2.ShardData{
//		{
//			ShardId: 1,
//			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
//				{
//					ReceiverShardId: 0,
//					SenderShardId:   0,
//				},
//			},
//		},
//	}
//
//	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityNilInvalidRecvShardIdOnShardedDataShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//	hdr.ShardInfo = []block2.ShardData{
//		{
//			ShardId: 0,
//			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
//				{
//					ReceiverShardId: 1,
//					SenderShardId:   0,
//				},
//			},
//		},
//	}
//
//	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityNilInvalidSenderShardIdOnShardedDataShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//	hdr.ShardInfo = []block2.ShardData{
//		{
//			ShardId: 0,
//			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
//				{
//					ReceiverShardId: 0,
//					SenderShardId:   1,
//				},
//			},
//		},
//	}
//
//	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityNilHeaderShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := &InterceptedMetaHeader{MetaBlock: nil}
//
//	assert.Equal(t, process.ErrNilMetaBlockHeader, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityOkValsShouldWork(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//	hdr.ShardInfo = []block2.ShardData{
//		{
//			ShardId: 0,
//			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
//				{
//					ReceiverShardId: 0,
//					SenderShardId:   0,
//				},
//			},
//		},
//	}
//
//	assert.Nil(t, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityOkValsWithEmptyShardDataShouldWork(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//
//	assert.Nil(t, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := &InterceptedMetaHeader{MetaBlock: &block2.MetaBlock{}}
//
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = nil
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//
//	assert.Equal(t, process.ErrNilPubKeysBitmap, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityAndValidityNilChronologyValidatorShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdr := NewInterceptedMetaHeader(
//		mock.NewMultiSigner(),
//		nil,
//	)
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//
//	assert.Equal(t, process.ErrNilChronologyValidator, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//
//	assert.Nil(t, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
//}
//
//func TestInterceptedMetaHeader_VerifySigOkValsShouldWork(t *testing.T) {
//	t.Parallel()
//
//	hdr := createTestInterceptedMetaHeader()
//	hdr.PrevHash = make([]byte, 0)
//	hdr.PubKeysBitmap = make([]byte, 0)
//	hdr.Signature = make([]byte, 0)
//	hdr.RootHash = make([]byte, 0)
//	hdr.PrevRandSeed = make([]byte, 0)
//	hdr.RandSeed = make([]byte, 0)
//
//	assert.Nil(t, hdr.VerifySig())
//}
