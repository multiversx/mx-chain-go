package shard_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewIntermediateProcessorsContainerFactory_NilShardCoord(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
		dPool,
		&economics.EconomicsData{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
		dPool,
		&economics.EconomicsData{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		nil,
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
		dPool,
		&economics.EconomicsData{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilAdrConv(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
		dPool,
		&economics.EconomicsData{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilStorer(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		nil,
		dPool,
		&economics.EconomicsData{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
		dPool,
		&economics.EconomicsData{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)
}

func TestIntermediateProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
		dPool,
		&economics.EconomicsData{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)

	container, err := ipcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 2, container.Len())
}
