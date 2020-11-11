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

func TestNewInterceptedShardHeaderDataFactory_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedShardHeaderDataFactory(nil)

	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.IntMarsh = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilSignMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.TxMarsh = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilHeaderSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)
	arg.HeaderSigVerifier = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilHeaderSigVerifier, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.Hash = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedShardHeaderDataFactory_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)
	arg.ValidityAttester = nil

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.True(t, check.IfNil(imh))
	assert.Equal(t, process.ErrNilValidityAttester, err)
}

func TestInterceptedShardHeaderDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)

	imh, err := NewInterceptedShardHeaderDataFactory(arg)
	assert.False(t, check.IfNil(imh))
	assert.Nil(t, err)
	assert.False(t, imh.IsInterfaceNil())

	marshalizer := &mock.MarshalizerMock{}
	emptyBlockHeader := &block.Header{}
	emptyBlockHeaderBuff, _ := marshalizer.Marshal(emptyBlockHeader)
	interceptedData, err := imh.Create(emptyBlockHeaderBuff)
	assert.Nil(t, err)

	_, ok := interceptedData.(*interceptedBlocks.InterceptedHeader)
	assert.True(t, ok)
}
