package interceptedBlocks_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultMetaArgument() *interceptedBlocks.ArgInterceptedBlockHeader {
	arg := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator:        mock.NewOneShardCoordinatorMock(),
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
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}

	hdr := createMockMetaHeader()
	arg.HdrBuff, _ = testMarshalizer.Marshal(hdr)

	return arg
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

//------- TestNewInterceptedHeader

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
			ShardID:               badShardId,
			HeaderHash:            nil,
			ShardMiniBlockHeaders: nil,
			TxCount:               0,
		},
	}
	buff, _ := testMarshalizer.Marshal(hdr)

	arg := createDefaultMetaArgument()
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

//------- getters

func TestInterceptedMetaHeader_Getters(t *testing.T) {
	t.Parallel()

	arg := createDefaultMetaArgument()
	inHdr, _ := interceptedBlocks.NewInterceptedMetaHeader(arg)

	hash := testHasher.Compute(string(arg.HdrBuff))

	assert.Equal(t, hash, inHdr.Hash())
	assert.True(t, inHdr.IsForCurrentShard())
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

//------- IsInterfaceNil

func TestInterceptedMetaHeader_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var inHdr *interceptedBlocks.InterceptedMetaHeader

	assert.True(t, check.IfNil(inHdr))
}
