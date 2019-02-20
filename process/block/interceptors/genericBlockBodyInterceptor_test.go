package interceptors_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewGenericBlockBodyInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	gbbi, err := interceptors.NewGenericBlockBodyInterceptor(
		nil,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, gbbi)
}

func TestNewGenericBlockBodyInterceptor_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}

	gbbi, err := interceptors.NewGenericBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		nil,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilCacher, err)
	assert.Nil(t, gbbi)
}

func TestNewGenericBlockBodyInterceptor_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}

	gbbi, err := interceptors.NewGenericBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilBlockBodyStorage, err)
	assert.Nil(t, gbbi)
}

func TestNewGenericBlockBodyInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	gbbi, err := interceptors.NewGenericBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		nil,
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, gbbi)
}

func TestNewGenericBlockBodyInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	gbbi, err := interceptors.NewGenericBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, gbbi)
}

func TestNewGenericBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	gbbi, err := interceptors.NewGenericBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, gbbi)
}
