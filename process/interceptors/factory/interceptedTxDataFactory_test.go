package factory

import (
	"testing"

	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedTxDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedTxDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedTxDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedTxDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedTxDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedTxDataFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.KeyGen = nil

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptedTxDataFactory_NilAdrConvShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.AddrConv = nil

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptedTxDataFactory_NilSignerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Signer = nil

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptedTxDataFactory_NilEconomicsFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.FeeHandler = nil

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestInterceptedTxDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()

	imh, err := NewInterceptedTxDataFactory(arg)
	assert.NotNil(t, imh)
	assert.Nil(t, err)

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &dataTransaction.Transaction{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)
	interceptedData, err := imh.Create(emptyTxBuff)

	_, ok := interceptedData.(*transaction.InterceptedTransaction)
	assert.True(t, ok)
}
