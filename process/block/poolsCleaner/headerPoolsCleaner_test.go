package poolsCleaner_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/poolsCleaner"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewHeaderPoolsCleaner_ShouldErrNilForkDetector(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		nil,
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherMock{},
		&mock.CacherMock{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldErrNilHeadersNoncesDataPool(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
		&mock.CacherMock{},
		&mock.CacherMock{},
	)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldErrNilHeadersDataPool(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		nil,
		&mock.CacherMock{},
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldErrNilNotarizedHeadersDataPool(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilNotarizedHeadersDataPool, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldWork(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherMock{},
		&mock.CacherMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, hpc)
}
