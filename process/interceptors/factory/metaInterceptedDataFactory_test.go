package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
)

func TestNewMetaInterceptedDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	midf, err := factory.NewMetaInterceptedDataFactory(nil, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewMetaInterceptedDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMetaInterceptedDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewMetaInterceptedDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMetaInterceptedDataFactory_NilMultiSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.MultiSigVerifier = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewMetaInterceptedDataFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.NodesCoordinator = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewMetaInterceptedDataFactory_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.FeeHandler = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewMetaInterceptedDataFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.KeyGen = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewMetaInterceptedDataFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Signer = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewMetaInterceptedDataFactory_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.AddrConv = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewMetaInterceptedDataFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	midf, err := factory.NewMetaInterceptedDataFactory(createMockArgument(), factory.InterceptedShardHeader)

	assert.False(t, check.IfNil(midf))
	assert.Nil(t, err)
}

//------- Create

func TestMetaInterceptedDataFactory_CreateUnknownDataTypeShouldErr(t *testing.T) {
	t.Parallel()

	undefinedDataType := factory.InterceptedDataType("undefined data type")
	midf, _ := factory.NewMetaInterceptedDataFactory(createMockArgument(), undefinedDataType)

	instance, err := midf.Create([]byte("buffer"))

	assert.Nil(t, instance)
	assert.Equal(t, process.ErrInterceptedDataTypeNotDefined, err)
}

func TestMetaInterceptedDataFactory_CreateInterceptedShardHdrShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyHdr := &block.Header{}
	emptyHdrBuff, _ := marshalizer.Marshal(emptyHdr)
	midf, _ := factory.NewMetaInterceptedDataFactory(createMockArgument(), factory.InterceptedShardHeader)

	instance, err := midf.Create(emptyHdrBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*interceptedBlocks.InterceptedHeader)
	assert.True(t, ok)
}

func TestMetaInterceptedDataFactory_CreateInterceptedMetaHdrShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyHdr := &block.Header{}
	emptyHdrBuff, _ := marshalizer.Marshal(emptyHdr)
	midf, _ := factory.NewMetaInterceptedDataFactory(createMockArgument(), factory.InterceptedMetaHeader)

	instance, err := midf.Create(emptyHdrBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*interceptedBlocks.InterceptedMetaHeader)
	assert.True(t, ok)
}

func TestMetaInterceptedDataFactory_CreateInterceptedTxShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &transaction.Transaction{}
	emptyHdrBuff, _ := marshalizer.Marshal(emptyTx)
	midf, _ := factory.NewMetaInterceptedDataFactory(createMockArgument(), factory.InterceptedTx)

	instance, err := midf.Create(emptyHdrBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*processTransaction.InterceptedTransaction)
	assert.True(t, ok)
}

//------- IsInterfaceNil

func TestMetaInterceptedDataFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	midf, _ := factory.NewMetaInterceptedDataFactory(createMockArgument(), factory.InterceptedShardHeader)
	midf = nil

	assert.True(t, check.IfNil(midf))
}
