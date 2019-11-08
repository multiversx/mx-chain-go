package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedTxBlockBodyDataFactory_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedTxBlockBodyDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedTxBlockBodyDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil

	imh, err := NewInterceptedTxBlockBodyDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedTxBlockBodyDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	imh, err := NewInterceptedTxBlockBodyDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedTxBlockBodyDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedTxBlockBodyDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestInterceptedTxBlockBodyDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()

	imh, err := NewInterceptedTxBlockBodyDataFactory(arg)
	assert.NotNil(t, imh)
	assert.Nil(t, err)

	marshalizer := &mock.MarshalizerMock{}
	emptyBlockBody := &block.Body{}
	emptyBlockBodyBuff, _ := marshalizer.Marshal(emptyBlockBody)
	interceptedData, err := imh.Create(emptyBlockBodyBuff)

	_, ok := interceptedData.(*interceptedBlocks.InterceptedTxBlockBody)
	assert.True(t, ok)
}
