package process_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetShardHeaderShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetShardHeader(hash, nil, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderShouldErrNilMarshalizer(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetShardHeader(hash, cacher, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetShardHeader(hash, cacher, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetShardHeaderShouldGetHeaderFromPool(t *testing.T) {
	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return hdr, true
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, _ := process.GetShardHeader(hash, cacher, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderShouldGetHeaderFromStorage(t *testing.T) {
	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, hash) {
						return marshalizer.Marshal(hdr)
					}
					return nil, errors.New("error")
				},
			}
		},
	}

	header, _ := process.GetShardHeader(hash, cacher, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderFromPoolShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	header, err := process.GetShardHeaderFromPool(hash, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderFromPoolShouldErrMissingHeader(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}

	header, err := process.GetShardHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromPoolShouldErrWrongTypeAssertion(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return &block.MetaBlock{}, true
		},
	}

	header, err := process.GetShardHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestGetShardHeaderFromPoolShouldWork(t *testing.T) {
	hash := []byte("X")

	hdr := &block.Header{Nonce: 10}
	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return hdr, true
		},
	}

	header, err := process.GetShardHeaderFromPool(hash, cacher)
	assert.Nil(t, err)
	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderFromStorageShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	storageService := &mock.ChainStorerMock{}

	header, err := process.GetShardHeaderFromStorage(hash, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderFromStorageShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetShardHeaderFromStorageShouldErrNilHeadersStorage(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilHeadersStorage, err)
}

func TestGetShardHeaderFromStorageShouldErrMissingHeader(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromStorageShouldErrUnmarshalWithoutSuccess(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, nil
				},
			}
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
}

func TestGetShardHeaderFromStorageShouldWork(t *testing.T) {
	hash := []byte("X")

	hdr := &block.Header{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, hash) {
						return marshalizer.Marshal(hdr)
					}
					return nil, errors.New("error")
				},
			}
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, err)
	assert.Equal(t, hdr, header)
}
