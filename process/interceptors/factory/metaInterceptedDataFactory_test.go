package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createMockMetaArgument() *factory.ArgMetaInterceptedDataFactory {
	return &factory.ArgMetaInterceptedDataFactory{
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              mock.HasherMock{},
		ShardCoordinator:    mock.NewOneShardCoordinatorMock(),
		MultiSigVerifier:    mock.NewMultiSigner(),
		ChronologyValidator: &mock.ChronologyValidatorStub{},
	}
}

func TestNewMetaInterceptedDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	midf, err := factory.NewMetaInterceptedDataFactory(nil, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewMetaInterceptedDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMetaArgument()
	arg.Marshalizer = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMetaInterceptedDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMetaArgument()
	arg.Hasher = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewMetaInterceptedDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMetaArgument()
	arg.ShardCoordinator = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMetaInterceptedDataFactory_NilMultiSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMetaArgument()
	arg.MultiSigVerifier = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewMetaInterceptedDataFactory_NilChronologyValidatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMetaArgument()
	arg.ChronologyValidator = nil

	midf, err := factory.NewMetaInterceptedDataFactory(arg, factory.InterceptedShardHeader)

	assert.Nil(t, midf)
	assert.Equal(t, process.ErrNilChronologyValidator, err)
}

func TestNewMetaInterceptedDataFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	midf, err := factory.NewMetaInterceptedDataFactory(createMockMetaArgument(), factory.InterceptedShardHeader)

	assert.False(t, check.IfNil(midf))
	assert.Nil(t, err)
}

//------- Create

func TestMetaInterceptedDataFactory_CreateUnknownDataTypeShouldErr(t *testing.T) {
	t.Parallel()

	undefinedDataType := factory.InterceptedDataType("undefined data type")
	midf, _ := factory.NewMetaInterceptedDataFactory(createMockMetaArgument(), undefinedDataType)

	instance, err := midf.Create([]byte("buffer"))

	assert.Nil(t, instance)
	assert.Equal(t, process.ErrInterceptedDataTypeNotDefined, err)
}

func TestMetaInterceptedDataFactory_CreateInterceptedShardHdrShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	emptyHdr := &block.Header{}
	emptyHdrBuff, _ := marshalizer.Marshal(emptyHdr)
	midf, _ := factory.NewMetaInterceptedDataFactory(createMockMetaArgument(), factory.InterceptedShardHeader)

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
	midf, _ := factory.NewMetaInterceptedDataFactory(createMockMetaArgument(), factory.InterceptedMetaHeader)

	instance, err := midf.Create(emptyHdrBuff)

	assert.NotNil(t, instance)
	assert.Nil(t, err)
	_, ok := instance.(*interceptedBlocks.InterceptedMetaHeader)
	assert.True(t, ok)
}

//------- IsInterfaceNil

func TestMetaInterceptedDataFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	midf, _ := factory.NewMetaInterceptedDataFactory(createMockMetaArgument(), factory.InterceptedShardHeader)
	midf = nil

	assert.True(t, check.IfNil(midf))
}
