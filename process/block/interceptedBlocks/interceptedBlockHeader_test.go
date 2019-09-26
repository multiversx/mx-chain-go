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

func createDefaultArgument() *interceptedBlocks.ArgInterceptedBlockHeader {
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

	hdr := createMockHeader()
	arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

	return arg
}

func createMockHeader() *dataBlock.Header {
	return &dataBlock.Header{
		Nonce:            hdrNonce,
		PrevHash:         []byte("prev hash"),
		PrevRandSeed:     []byte("prev rand seed"),
		RandSeed:         []byte("rand seed"),
		PubKeysBitmap:    []byte("bitmap"),
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

func TestNewInterceptedHeader_NilHdrShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.HdrBuff = nil

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedHeader_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.Marshalizer = nil

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedHeader_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.Hasher = nil

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedHeader_NilMultiSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.MultiSigVerifier = nil

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptedHeader_NilChronologyValidatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.ChronologyValidator = nil

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilChronologyValidator, err)
}

func TestNewInterceptedHeader_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.ShardCoordinator = nil

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedHeader_MarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.HdrBuff = []byte("invalid buffer")

	inHdr, err := interceptedBlocks.NewInterceptedHeader(arg)

	assert.Nil(t, inHdr)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestNewInterceptedHeader_NotForThisShardShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
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

	arg := createDefaultArgument()
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

	arg := createDefaultArgument()
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

	hdr := createMockHeader()
	hdr.PubKeysBitmap = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestInterceptedHeader_CheckValidityNilPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockHeader()
	hdr.PrevHash = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilPreviousBlockHash, err)
}

func TestInterceptedHeader_CheckValidityNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockHeader()
	hdr.Signature = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilSignature, err)
}

func TestInterceptedHeader_CheckValidityNilRootHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockHeader()
	hdr.RootHash = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilRootHash, err)
}

func TestInterceptedHeader_CheckValidityNilRandSeedShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockHeader()
	hdr.RandSeed = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilRandSeed, err)
}

func TestInterceptedHeader_CheckValidityNilPrevRandSeedShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createMockHeader()
	hdr.PrevRandSeed = nil
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultArgument()
	arg.HdrBuff = buff
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Equal(t, process.ErrNilPrevRandSeed, err)
}

func TestInterceptedHeader_CheckValidityShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedHeader(arg)

	err := inHdr.CheckValidity()

	assert.Nil(t, err)
}

//------- getters

func TestInterceptedHeader_Getters(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
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
