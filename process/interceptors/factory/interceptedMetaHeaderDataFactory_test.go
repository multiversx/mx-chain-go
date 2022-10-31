package factory

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/versioning"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	processMocks "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

var errSingleSignKeyGenMock = errors.New("errSingleSignKeyGenMock")
var errSignerMockVerifySigFails = errors.New("errSignerMockVerifySigFails")
var sigOk = []byte("signature")

func createMockKeyGen() crypto.KeyGenerator {
	return &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, e error) {
			if string(b) == "" {
				return nil, errSingleSignKeyGenMock
			}

			return &mock.SingleSignPublicKey{}, nil
		},
	}
}

func createMockSigner() crypto.SingleSigner {
	return &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			if !bytes.Equal(sig, sigOk) {
				return errSignerMockVerifySigFails
			}
			return nil
		},
	}
}

func createMockPubkeyConverter() core.PubkeyConverter {
	return mock.NewPubkeyConverterMock(32)
}

func createMockFeeHandler() process.FeeHandler {
	return &mock.FeeHandlerStub{}
}

func createMockComponentHolders() (*mock.CoreComponentsMock, *mock.CryptoComponentsMock) {
	coreComponents := &mock.CoreComponentsMock{
		IntMarsh:            &mock.MarshalizerMock{},
		TxMarsh:             &mock.MarshalizerMock{},
		Hash:                &hashingMocks.HasherMock{},
		TxSignHasherField:   &hashingMocks.HasherMock{},
		UInt64ByteSliceConv: mock.NewNonceHashConverterMock(),
		AddrPubKeyConv:      createMockPubkeyConverter(),
		ChainIdCalled: func() string {
			return "chainID"
		},
		TxVersionCheckField:        versioning.NewTxVersionChecker(1),
		EpochNotifierField:         &epochNotifier.EpochNotifierStub{},
		HardforkTriggerPubKeyField: []byte("provided hardfork pub key"),
		EnableEpochsHandlerField:   &testscommon.EnableEpochsHandlerStub{},
	}
	cryptoComponents := &mock.CryptoComponentsMock{
		BlockSig:          createMockSigner(),
		TxSig:             createMockSigner(),
		MultiSigContainer: cryptoMocks.NewMultiSignerContainerMock(cryptoMocks.NewMultiSigner()),
		BlKeyGen:          createMockKeyGen(),
		TxKeyGen:          createMockKeyGen(),
	}

	return coreComponents, cryptoComponents
}

func createMockArgument(
	coreComponents *mock.CoreComponentsMock,
	cryptoComponents *mock.CryptoComponentsMock,
) *ArgInterceptedDataFactory {
	return &ArgInterceptedDataFactory{
		CoreComponents:               coreComponents,
		CryptoComponents:             cryptoComponents,
		ShardCoordinator:             mock.NewOneShardCoordinatorMock(),
		NodesCoordinator:             shardingMocks.NewNodesCoordinatorMock(),
		FeeHandler:                   createMockFeeHandler(),
		WhiteListerVerifiedTxs:       &testscommon.WhiteListHandlerStub{},
		HeaderSigVerifier:            &mock.HeaderSigVerifierStub{},
		ValidityAttester:             &mock.ValidityAttesterStub{},
		HeaderIntegrityVerifier:      &mock.HeaderIntegrityVerifierStub{},
		EpochStartTrigger:            &mock.EpochStartTriggerStub{},
		ArgsParser:                   &mock.ArgumentParserMock{},
		PeerSignatureHandler:         &processMocks.PeerSignatureHandlerStub{},
		SignaturesHandler:            &processMocks.SignaturesHandlerStub{},
		HeartbeatExpiryTimespanInSec: 30,
		PeerID:                       "pid",
	}
}

func TestNewInterceptedMetaHeaderDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedMetaHeaderDataFactory(nil)

	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.IntMarsh = nil
	arg := createMockArgument(coreComp, cryptoComp)

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilSignMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.TxMarsh = nil
	arg := createMockArgument(coreComp, cryptoComp)

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.Hash = nil
	arg := createMockArgument(coreComp, cryptoComp)

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilHeaderSigVerifierShouldErr(t *testing.T) {
	t.Parallel()
	coreComp, cryptoComp := createMockComponentHolders()
	arg := createMockArgument(coreComp, cryptoComp)
	arg.HeaderSigVerifier = nil

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilHeaderSigVerifier, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilHeaderIntegrityVerifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	arg := createMockArgument(coreComp, cryptoComp)
	arg.HeaderIntegrityVerifier = nil

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilHeaderIntegrityVerifier, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	arg := createMockArgument(coreComp, cryptoComp)
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilChainIdShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.ChainIdCalled = func() string {
		return ""
	}
	arg := createMockArgument(coreComp, cryptoComp)

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestNewInterceptedMetaHeaderDataFactory_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	arg := createMockArgument(coreComp, cryptoComp)
	arg.ValidityAttester = nil

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilValidityAttester, err)
}

func TestNewInterceptedMetaHeaderDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	arg := createMockArgument(coreComp, cryptoComp)

	imh, err := NewInterceptedMetaHeaderDataFactory(arg)
	assert.False(t, check.IfNil(imh))
	assert.Nil(t, err)
	assert.False(t, imh.IsInterfaceNil())

	marshalizer := &mock.MarshalizerMock{}
	emptyMetaHeader := &block.Header{}
	emptyMetaHeaderBuff, _ := marshalizer.Marshal(emptyMetaHeader)
	interceptedData, err := imh.Create(emptyMetaHeaderBuff)
	assert.Nil(t, err)

	_, ok := interceptedData.(*interceptedBlocks.InterceptedMetaHeader)
	assert.True(t, ok)
}
