package metachain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
)

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestNewIntermediateProcessorsContainerFactory_NilShardCoord(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		nil,
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		nil,
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		nil,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilAdrConv(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		nil,
		&mock.ChainStorerMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilStorer(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		nil,
		dataRetrieverMock.NewPoolsHolderMock(),
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilPoolsHolder(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		nil,
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilEconomicsFeeHandler(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		nil,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewIntermediateProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)
	assert.False(t, ipcf.IsInterfaceNil())
}

func TestIntermediateProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	ipcf, err := metachain.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)

	container, err := ipcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 2, container.Len())
}
