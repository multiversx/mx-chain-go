package factory_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
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

func createMockArgument() *factory.ArgInterceptedDataFactory {
	return &factory.ArgInterceptedDataFactory{
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           mock.HasherMock{},
		KeyGen:           createMockKeyGen(),
		Signer:           createMockSigner(),
		AddrConv:         createMockAddressConverter(),
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
	}
}

func TestNewInterceptedDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	idf, err := factory.NewInterceptedDataFactory(nil, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil
	idf, err := factory.NewInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil
	idf, err := factory.NewInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedDataFactory_NilKeygenShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.KeyGen = nil
	idf, err := factory.NewInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptedDataFactory_NilSignerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Signer = nil
	idf, err := factory.NewInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptedDataFactory_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.AddrConv = nil
	idf, err := factory.NewInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptedDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil
	idf, err := factory.NewInterceptedDataFactory(arg, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedDataFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	idf, err := factory.NewInterceptedDataFactory(createMockArgument(), factory.InterceptedTx)

	assert.False(t, check.IfNil(idf))
	assert.Nil(t, err)
}

//------- Create

func TestInterceptedDataFactory_CreateUnknownDataTypeShouldErr(t *testing.T) {
	t.Parallel()

	undefinedDataType := factory.InterceptedDataType("undefined data type")
	idf, _ := factory.NewInterceptedDataFactory(createMockArgument(), undefinedDataType)

	instance, err := idf.Create([]byte("buffer"))

	assert.Nil(t, instance)
	assert.Equal(t, process.ErrInterceptedDataTypeNotDefined, err)
}

func TestInterceptedDataFactory_CreateInterceptedTxShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &dataTransaction.Transaction{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)

	idf, _ := factory.NewInterceptedDataFactory(createMockArgument(), factory.InterceptedTx)

	instance, err := idf.Create(emptyTxBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*transaction.InterceptedTransaction)
	assert.True(t, ok)
}

//------- IsInterfaceNil

func TestInterceptedDataFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	idf, _ := factory.NewInterceptedDataFactory(createMockArgument(), factory.InterceptedTx)
	idf = nil

	assert.True(t, check.IfNil(idf))
}
