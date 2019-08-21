package factory_test

import (
	"bytes"
	"errors"
	"testing"

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

func TestNewInterceptedDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	idf, err := factory.NewInterceptedDataFactory(nil, factory.InterceptedTx)

	assert.Nil(t, idf)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedDataFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	idf, err := factory.NewInterceptedDataFactory(&factory.InterceptedDataFactoryArgument{}, factory.InterceptedTx)

	assert.NotNil(t, idf)
	assert.Nil(t, err)
}

//------- Create

func TestInterceptedDataFactory_CreateUnknownDataTypeShouldErr(t *testing.T) {
	t.Parallel()

	undefinedDataType := factory.InterceptedDataType("undefined data type")
	idf, _ := factory.NewInterceptedDataFactory(&factory.InterceptedDataFactoryArgument{}, undefinedDataType)

	instance, err := idf.Create([]byte("buffer"))

	assert.Nil(t, instance)
	assert.Equal(t, process.ErrInterceptedDataTypeNotDefined, err)
}

func TestInterceptedDataFactory_CreateInterceptedTxShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &dataTransaction.Transaction{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)

	argument := &factory.InterceptedDataFactoryArgument{
		Marshalizer:      marshalizer,
		Hasher:           mock.HasherMock{},
		KeyGen:           createMockKeyGen(),
		Signer:           createMockSigner(),
		AddrConv:         createMockAddressConverter(),
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
	}
	idf, _ := factory.NewInterceptedDataFactory(argument, factory.InterceptedTx)

	instance, err := idf.Create(emptyTxBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*transaction.InterceptedTransaction)
	assert.True(t, ok)
}
