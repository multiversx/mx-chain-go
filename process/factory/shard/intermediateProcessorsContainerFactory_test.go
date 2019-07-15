package shard

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewIntermediateProcessorsContainerFactory_NilShardCoord(t *testing.T) {
	t.Parallel()

	ipcf, err := NewIntermediateProcessorsContainerFactory(
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	ipcf, err := NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	ipcf, err := NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		nil,
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilAdrConv(t *testing.T) {
	t.Parallel()

	ipcf, err := NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilStorer(t *testing.T) {
	t.Parallel()

	ipcf, err := NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		nil,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	ipcf, err := NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)
}

func TestIntermediateProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	ipcf, err := NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.SpecialAddressHandlerMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)

	container, err := ipcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 2, container.Len())
}
