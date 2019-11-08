package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedShardHeaderDataFactory_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedShardHeaderDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilMultiSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.MultiSigVerifier = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.NodesCoordinator = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestInterceptedShardHeaderDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.NotNil(t, imh)
	assert.Nil(t, err)

	marshalizer := &mock.MarshalizerMock{}
	emptyBlockHeader := &block.Header{}
	emptyBlockHeaderBuff, _ := marshalizer.Marshal(emptyBlockHeader)
	interceptedData, err := imh.Create(emptyBlockHeaderBuff)

	_, ok := interceptedData.(*interceptedBlocks.InterceptedHeader)
	assert.True(t, ok)
}
