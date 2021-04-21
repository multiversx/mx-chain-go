package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedUnsignedTxDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedUnsignedTxDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.IntMarsh = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilSignMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.TxMarsh = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.Hash = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.AddrPubKeyConv = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestInterceptedUnsignedTxDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.NotNil(t, imh)
	assert.Nil(t, err)
	assert.False(t, imh.IsInterfaceNil())

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &smartContractResult.SmartContractResult{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)
	interceptedData, err := imh.Create(emptyTxBuff)
	assert.Nil(t, err)

	_, ok := interceptedData.(*unsigned.InterceptedUnsignedTransaction)
	assert.True(t, ok)
}
