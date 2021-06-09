package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedMiniblockDataFactory_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedMiniblockDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedMiniblockDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.IntMarsh = nil
	arg := createMockArgument(coreComp, cryptoComp)

	imdf, err := NewInterceptedMiniblockDataFactory(arg)
	assert.True(t, check.IfNil(imdf))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedMiniblockDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.Hash = nil
	arg := createMockArgument(coreComp, cryptoComp)

	imdf, err := NewInterceptedMiniblockDataFactory(arg)
	assert.True(t, check.IfNil(imdf))
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedMiniblockDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	arg := createMockArgument(coreComp, cryptoComp)
	arg.ShardCoordinator = nil

	imdf, err := NewInterceptedMiniblockDataFactory(arg)
	assert.True(t, check.IfNil(imdf))
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestInterceptedMiniblockDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	arg := createMockArgument(coreComp, cryptoComp)

	imdf, err := NewInterceptedMiniblockDataFactory(arg)
	assert.False(t, check.IfNil(imdf))
	assert.Nil(t, err)

	marshalizer := &mock.MarshalizerMock{}
	emptyBlockBody := &block.Body{}
	emptyBlockBodyBuff, _ := marshalizer.Marshal(emptyBlockBody)
	interceptedData, err := imdf.Create(emptyBlockBodyBuff)
	assert.Nil(t, err)

	_, ok := interceptedData.(*interceptedBlocks.InterceptedMiniblock)
	assert.True(t, ok)
}
