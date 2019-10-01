package factory_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
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

func createMockAddressConverter() state.AddressConverter {
	return &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return mock.NewAddressMock(pubKey), nil
		},
	}
}

func createMockShardArgument() *factory.ArgShardInterceptedDataFactory {
	return &factory.ArgShardInterceptedDataFactory{
		ArgMetaInterceptedDataFactory: createMockMetaArgument(),
		KeyGen:                        createMockKeyGen(),
		Signer:                        createMockSigner(),
		AddrConv:                      createMockAddressConverter(),
	}
}

func TestNewShardInterceptedDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	sidf, err := factory.NewShardInterceptedDataFactory(nil, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewShardInterceptedDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.Marshalizer = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardInterceptedDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.Hasher = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewShardInterceptedDataFactory_NilKeygenShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.KeyGen = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewShardInterceptedDataFactory_NilSignerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.Signer = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewShardInterceptedDataFactory_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.AddrConv = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewShardInterceptedDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.ShardCoordinator = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewShardInterceptedDataFactory_NilMultiSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.MultiSigVerifier = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewShardInterceptedDataFactory_NilChronologyValidatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockShardArgument()
	arg.ChronologyValidator = nil

	sidf, err := factory.NewShardInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, sidf)
	assert.Equal(t, process.ErrNilChronologyValidator, err)
}

func TestNewShardInterceptedDataFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	sidf, err := factory.NewShardInterceptedDataFactory(createMockShardArgument(), factory.InterceptedTx)

	assert.False(t, check.IfNil(sidf))
	assert.Nil(t, err)
}

//------- Create

func TestShardInterceptedDataFactory_CreateUnknownDataTypeShouldErr(t *testing.T) {
	t.Parallel()

	undefinedDataType := factory.InterceptedDataType("undefined data type")
	sidf, _ := factory.NewShardInterceptedDataFactory(createMockShardArgument(), undefinedDataType)

	instance, err := sidf.Create([]byte("buffer"))

	assert.Nil(t, instance)
	assert.Equal(t, process.ErrInterceptedDataTypeNotDefined, err)
}

func TestShardInterceptedDataFactory_CreateInterceptedTxShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &dataTransaction.Transaction{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)
	sidf, _ := factory.NewShardInterceptedDataFactory(createMockShardArgument(), factory.InterceptedTx)

	instance, err := sidf.Create(emptyTxBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*transaction.InterceptedTransaction)
	assert.True(t, ok)
}

func TestShardInterceptedDataFactory_CreateInterceptedUnsignedTxShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &smartContractResult.SmartContractResult{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)
	idf, _ := factory.NewShardInterceptedDataFactory(createMockShardArgument(), factory.InterceptedUnsignedTx)

	instance, err := idf.Create(emptyTxBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*unsigned.InterceptedUnsignedTransaction)
	assert.True(t, ok)
}

func TestShardInterceptedDataFactory_CreateInterceptedShardHdrShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyHdr := &block.Header{}
	emptyHdrBuff, _ := marshalizer.Marshal(emptyHdr)
	sidf, _ := factory.NewShardInterceptedDataFactory(createMockShardArgument(), factory.InterceptedShardHeader)

	instance, err := sidf.Create(emptyHdrBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*interceptedBlocks.InterceptedHeader)
	assert.True(t, ok)
}

func TestShardInterceptedDataFactory_CreateInterceptedMetaHdrShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyHdr := &block.Header{}
	emptyHdrBuff, _ := marshalizer.Marshal(emptyHdr)
	midf, _ := factory.NewShardInterceptedDataFactory(createMockShardArgument(), factory.InterceptedMetaHeader)

	instance, err := midf.Create(emptyHdrBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*interceptedBlocks.InterceptedMetaHeader)
	assert.True(t, ok)
}

//------- IsInterfaceNil

func TestShardInterceptedDataFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sidf, _ := factory.NewShardInterceptedDataFactory(createMockShardArgument(), factory.InterceptedTx)
	sidf = nil

	assert.True(t, check.IfNil(sidf))
}
