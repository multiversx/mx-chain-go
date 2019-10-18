package interceptedBlocks_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &mock.MarshalizerMock{}
var testHasher = mock.HasherMock{}
var hdrNonce = uint64(56)
var hdrShardId = uint32(1)
var hdrRound = uint64(67)
var hdrEpoch = uint32(78)

func createDefaultShardArgument() *interceptedBlocks.ArgInterceptedBlockHeader {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		MultiSigVerifier: mock.NewMultiSigner(),
		Hasher:           testHasher,
		Marshalizer:      testMarshalizer,
		NodesCoordinator: &mock.NodesCoordinatorMock{},
	}

	hdr := createMockShardHeader()
	arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

	return arg
}

func createMockShardHeader() *dataBlock.Header {
	return &dataBlock.Header{
		Nonce:            hdrNonce,
		PrevHash:         []byte("prev hash"),
		PrevRandSeed:     []byte("prev rand seed"),
		RandSeed:         []byte("rand seed"),
		PubKeysBitmap:    []byte{1},
		ShardId:          hdrShardId,
		TimeStamp:        0,
		Round:            hdrRound,
		Epoch:            hdrEpoch,
		BlockBodyType:    dataBlock.TxBlock,
		Signature:        []byte("signature"),
		MiniBlockHeaders: nil,
		PeerChanges:      nil,
		RootHash:         []byte("root hash"),
		MetaBlockHashes:  nil,
		TxCount:          0,
	}
}

//------- TestNewInterceptedHeader

func TestNewInterceptedHeader_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	inHdr, err := interceptedBlocks.NewInterceptedHeader(nil)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedHeader_MarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultShardArgument()
	arg.HdrBuff = []byte("invalid buffer")

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestNewInterceptedHeader_NotForThisShardShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultShardArgument()
	arg.ShardCoordinator = &mock.CoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return hdrShardId + 2
		},
		SelfIdCalled: func() uint32 {
			return hdrShardId + 1
		},
	}

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.False(t, check.IfNil(inHdr))
	assert.Nil(t, err)
	assert.False(t, inHdr.IsForCurrentShard())
}

func TestNewInterceptedHeader_ForThisShardShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultShardArgument()
	arg.ShardCoordinator = &mock.CoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return hdrShardId + 2
		},
		SelfIdCalled: func() uint32 {
			return hdrShardId
		},
	}

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.False(t, check.IfNil(inHdr))
	assert.Nil(t, err)
	assert.True(t, inHdr.IsForCurrentShard())
}

func TestNewInterceptedHeader_MetachainForThisShardShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultShardArgument()
	arg.ShardCoordinator = &mock.CoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return hdrShardId + 2
		},
		SelfIdCalled: func() uint32 {
			return sharding.MetachainShardId
		},
	}

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.False(t, check.IfNil(inHdr))
	assert.Nil(t, err)
	assert.True(t, inHdr.IsForCurrentShard())
}

//------- CheckValidity

func TestInterceptedHeader_CheckValidityNilPubKeyBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockShardHeader()
	hdr.PubKeysBitmap = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultShardArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestInterceptedHeader_ErrorInMiniBlockShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockShardHeader()
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
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultShardArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedHeader_CheckValidityShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultShardArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Nil(t, err)
}

//------- getters

func TestInterceptedHeader_Getters(t *testing.T) {
	t.Parallel()

	arg := createDefaultShardArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	hash := testHasher.Compute(string(arg.HdrBuff))

	assert.Equal(t, hash, inHdr.Hash())
}

//------- IsInterfaceNil

func TestInterceptedHeader_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var inHdr *interceptedBlocks.InterceptedHeader

	assert.True(t, check.IfNil(inHdr))
}
