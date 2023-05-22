package process_test

import (
	"bytes"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShardHeaderShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}

	header, err := process.GetShardHeader(hash, nil, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{}
	storageService := &storageStubs.ChainStorerStub{}

	header, err := process.GetShardHeader(hash, cacher, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetShardHeader(hash, cacher, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetShardHeaderShouldGetHeaderFromPool(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return hdr, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}

	header, _ := process.GetShardHeader(hash, cacher, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderShouldGetHeaderFromStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, hash) {
						return marshalizer.Marshal(hdr)
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	header, _ := process.GetShardHeader(hash, cacher, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}

	header, err := process.GetMetaHeader(hash, nil, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{}
	storageService := &storageStubs.ChainStorerStub{}

	header, err := process.GetMetaHeader(hash, cacher, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMetaHeaderShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetMetaHeader(hash, cacher, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMetaHeaderShouldGetHeaderFromPool(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 1}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return hdr, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}

	header, _ := process.GetMetaHeader(hash, cacher, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderShouldGetHeaderFromStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 1}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, hash) {
						return marshalizer.Marshal(hdr)
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	header, _ := process.GetMetaHeader(hash, cacher, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderFromPoolShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	header, err := process.GetShardHeaderFromPool(hash, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderFromPoolShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}

	header, err := process.GetShardHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetShardHeaderFromPoolShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return &block.MetaBlock{}, nil
		},
	}

	header, err := process.GetShardHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestGetShardHeaderFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.Header{Nonce: 10}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return hdr, nil
		},
	}

	header, err := process.GetShardHeaderFromPool(hash, cacher)
	assert.Nil(t, err)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderFromPoolShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	header, err := process.GetMetaHeaderFromPool(hash, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderFromPoolShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}

	header, err := process.GetMetaHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetMetaHeaderFromPoolShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return &block.Header{}, nil
		},
	}

	header, err := process.GetMetaHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestGetMetaHeaderFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 10}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return hdr, nil
		},
	}

	header, err := process.GetMetaHeaderFromPool(hash, cacher)
	assert.Nil(t, err)
	assert.Equal(t, hdr, header)
}

func TestGetHeaderFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	shardHeader := &block.Header{Nonce: 42}
	metaHeader := &block.MetaBlock{Nonce: 43}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if unitType == dataRetriever.BlockHeaderUnit && bytes.Equal(key, []byte("shard")) {
						return marshalizer.Marshal(shardHeader)
					} else if unitType == dataRetriever.MetaBlockUnit && bytes.Equal(key, []byte("meta")) {
						return marshalizer.Marshal(metaHeader)
					}

					return nil, errors.New("error")
				},
			}, nil
		},
	}

	header, err := process.GetHeaderFromStorage(0, []byte("shard"), marshalizer, storageService)
	assert.Nil(t, err)
	assert.Equal(t, shardHeader, header)

	header, err = process.GetHeaderFromStorage(core.MetachainShardId, []byte("meta"), marshalizer, storageService)
	assert.Nil(t, err)
	assert.Equal(t, metaHeader, header)
}

func TestGetShardHeaderFromStorageShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	storageService := &storageStubs.ChainStorerStub{}

	header, err := process.GetShardHeaderFromStorage(hash, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderFromStorageShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetShardHeaderFromStorageShouldErrGetHeadersStorageReturnsErr(t *testing.T) {
	t.Parallel()
	hash := []byte("X")

	expectedErr := errors.New("expected error")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return nil, expectedErr
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, expectedErr, err)
}

func TestGetShardHeaderFromStorageShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetShardHeaderFromStorageShouldErrUnmarshalWithoutSuccess(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, nil
				},
			}, nil
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
}

func TestGetShardHeaderFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.Header{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, hash) {
						return marshalizer.Marshal(hdr)
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	header, err := process.GetShardHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, err)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderFromStorageShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	storageService := &storageStubs.ChainStorerStub{}

	header, err := process.GetMetaHeaderFromStorage(hash, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMetaHeaderFromStorageShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMetaHeaderFromStorageShouldErrGetHeadersStorageReturnsErr(t *testing.T) {
	t.Parallel()
	hash := []byte("X")

	expectedErr := errors.New("expected error")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return nil, expectedErr
		},
	}

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, expectedErr, err)
}

func TestGetMetaHeaderFromStorageShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetMetaHeaderFromStorageShouldErrUnmarshalWithoutSuccess(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, nil
				},
			}, nil
		},
	}

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
}

func TestGetMetaHeaderFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.MetaBlock{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, hash) {
						return marshalizer.Marshal(hdr)
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, err)
	assert.Equal(t, hdr, header)
}

func TestGetMarshalizedHeaderFromStorageShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	storageService := &storageStubs.ChainStorerStub{}

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, nil, storageService)
	assert.Nil(t, headerMarsh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMarshalizedHeaderFromStorageShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, nil)
	assert.Nil(t, headerMarsh)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMarshalizedHeaderFromStorageShouldErrGetHeadersStorageReturnsErr(t *testing.T) {
	t.Parallel()
	hash := []byte("X")

	expectedErr := errors.New("expected error")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return nil, expectedErr
		},
	}

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	assert.Nil(t, headerMarsh)
	assert.Equal(t, expectedErr, err)
}

func TestGetMarshalizedHeaderFromStorageShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	assert.Nil(t, headerMarsh)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetMarshalizedHeaderFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, hash) {
						return marshalizer.Marshal(hdr)
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}

	hdrMarsh, _ := marshalizer.Marshal(hdr)
	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	assert.Equal(t, hdrMarsh, headerMarsh)
	assert.Nil(t, err)
}

func TestGetShardHeaderWithNonceShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetShardHeaderWithNonce(
		nonce,
		shardId,
		nil,
		marshalizer,
		storageService,
		uint64Converter)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderWithNonceShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}
	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetShardHeaderWithNonce(
		nonce,
		shardId,
		cacher,
		nil,
		storageService,
		uint64Converter)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderWithNonceShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetShardHeaderWithNonce(
		nonce,
		shardId,
		cacher,
		marshalizer,
		nil,
		uint64Converter)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetShardHeaderWithNonceShouldErrNilUint64Converter(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}

	header, hash, err := process.GetShardHeaderWithNonce(
		nonce,
		shardId,
		cacher,
		marshalizer,
		storageService,
		nil)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilUint64Converter, err)
}

func TestGetShardHeaderWithNonceShouldGetHeaderFromPool(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	nonce := uint64(1)
	shardId := uint32(0)
	hdr := &block.Header{Nonce: nonce}

	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, _, _ := process.GetShardHeaderWithNonce(
		nonce,
		shardId,
		cacher,
		marshalizer,
		storageService,
		uint64Converter)

	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderWithNonceShouldGetHeaderFromStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	nonce := uint64(1)
	nonceToByte := []byte("1")
	shardId := uint32(0)
	hdr := &block.Header{Nonce: nonce}

	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}

	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, _, _ := process.GetShardHeaderWithNonce(
		nonce,
		shardId,
		cacher,
		marshalizer,
		storageService,
		uint64Converter)

	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderWithNonceShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetMetaHeaderWithNonce(
		nonce,
		nil,
		marshalizer,
		storageService,
		uint64Converter)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderWithNonceShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}
	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetMetaHeaderWithNonce(
		nonce,
		cacher,
		nil,
		storageService,
		uint64Converter)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMetaHeaderWithNonceShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetMetaHeaderWithNonce(
		nonce,
		cacher,
		marshalizer,
		nil,
		uint64Converter)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMetaHeaderWithNonceShouldErrNilUint64Converter(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}

	header, hash, err := process.GetMetaHeaderWithNonce(
		nonce,
		cacher,
		marshalizer,
		storageService,
		nil)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilUint64Converter, err)
}

func TestGetMetaHeaderWithNonceShouldGetHeaderFromPool(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	nonce := uint64(1)

	hdr := &block.MetaBlock{Nonce: nonce}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, _, _ := process.GetMetaHeaderWithNonce(
		nonce,
		cacher,
		marshalizer,
		storageService,
		uint64Converter)

	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderWithNonceShouldGetHeaderFromStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	nonce := uint64(1)
	nonceToByte := []byte("1")

	hdr := &block.MetaBlock{Nonce: nonce}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, _, _ := process.GetMetaHeaderWithNonce(
		nonce,
		cacher,
		marshalizer,
		storageService,
		uint64Converter)

	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderFromPoolWithNonceShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, nil)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenShardIdHashMapIsNil(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetShardHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenLoadFromShardIdHashMapFails(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetShardHeaderFromPoolWithNonceShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetShardHeaderFromPoolWithNonceShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return &block.MetaBlock{}, nil
		},
	}

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetShardHeaderFromPoolWithNonceShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	nonce := uint64(1)
	shardId := uint32(0)

	hdr := &block.Header{Nonce: nonce}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}

	header, headerHash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher)
	assert.Nil(t, err)
	assert.Equal(t, hash, headerHash)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrNilCacher(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, nil)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenShardIdHashMapIsNil(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenLoadFromShardIdHashMapFails(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{&block.Header{}}, [][]byte{hash}, nil
		},
	}

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestGetMetaHeaderFromPoolWithNonceShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	nonce := uint64(1)

	hdr := &block.MetaBlock{Nonce: nonce}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}

	header, headerHash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, err)
	assert.Equal(t, hash, headerHash)
	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	uint64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		nil,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrNilUint64Converter(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	storageService := &storageStubs.ChainStorerStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		nil,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilUint64Converter, err)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		uint64Converter,
		nil)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrGetHeadersStorageReturnsErr(t *testing.T) {
	t.Parallel()
	nonce := uint64(1)
	shardId := uint32(0)

	expectedErr := errors.New("expected error")
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return nil, expectedErr
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, expectedErr, err)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrMissingHashForHeaderNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)

	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrMissingHashForHeaderNonce, err)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, hash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetShardHeaderFromStorageWithNonceShouldErrUnmarshalWithoutSuccess(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return nil, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, hash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
}

func initDefaultStorageServiceAndConverter(nonce uint64, hash []byte, hdr data.HeaderHandler) (
	dataRetriever.StorageService,
	typeConverters.Uint64ByteSliceConverter,
) {
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	return storageService, uint64Converter
}

func TestGetHeaderFromStorageWithNonceShouldWorkForShard(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	hdr := &block.Header{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}

	storageService, uint64Converter := initDefaultStorageServiceAndConverter(nonce, hash, hdr)
	header, headerHash, err := process.GetHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, err)
	assert.Equal(t, hash, headerHash)
	assert.Equal(t, hdr, header)

}

func TestGetShardHeaderFromStorageWithNonceShouldWorkForShard(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	hdr := &block.Header{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}

	storageService, uint64Converter := initDefaultStorageServiceAndConverter(nonce, hash, hdr)
	header, headerHash, err := process.GetShardHeaderFromStorageWithNonce(
		nonce,
		shardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, err)
	assert.Equal(t, hash, headerHash)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	uint64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		nil,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrNilUint64Converter(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	storageService := &storageStubs.ChainStorerStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		storageService,
		nil,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilUint64Converter, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	storageService := &storageStubs.ChainStorerStub{}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}

	header, hash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		nil)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrGetHeadersStorageReturnsErr(t *testing.T) {
	t.Parallel()
	nonce := uint64(1)

	expectedErr := errors.New("expected error")
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return nil, expectedErr
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, expectedErr, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrMissingHashForHeaderNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerMock{}

	header, hash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrMissingHashForHeaderNonce, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, hash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrUnmarshalWithoutSuccess(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return nil, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, hash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	hdr := &block.MetaBlock{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, headerHash, err := process.GetMetaHeaderFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, err)
	assert.Equal(t, hash, headerHash)
	assert.Equal(t, hdr, header)
}

func TestGetHeaderFromStorageWithNonceShouldWorkForMeta(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	hdr := &block.MetaBlock{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}, nil
		},
	}
	uint64Converter := &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(n uint64) []byte {
			if n == nonce {
				return nonceToByte
			}

			return nil
		},
	}

	header, headerHash, err := process.GetHeaderFromStorageWithNonce(
		nonce,
		core.MetachainShardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, err)
	assert.Equal(t, hash, headerHash)
	assert.Equal(t, hdr, header)
}

func TestGetTransactionHandler_Errors(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	storageService := &storageStubs.ChainStorerStub{}
	marshaller := &mock.MarshalizerMock{}
	shardedDataCacherNotifier := testscommon.NewShardedDataStub()

	t.Run("errors if the sharded cacher is nil", func(t *testing.T) {
		t.Parallel()

		tx, err := process.GetTransactionHandler(
			0,
			0,
			hash,
			nil,
			storageService,
			marshaller,
			process.SearchMethodJustPeek)

		assert.Nil(t, tx)
		assert.Equal(t, process.ErrNilShardedDataCacherNotifier, err)
	})
	t.Run("errors if the storage service is nil", func(t *testing.T) {
		t.Parallel()

		tx, err := process.GetTransactionHandler(
			0,
			0,
			hash,
			shardedDataCacherNotifier,
			nil,
			marshaller,
			process.SearchMethodJustPeek)

		assert.Nil(t, tx)
		assert.Equal(t, process.ErrNilStorage, err)
	})
	t.Run("errors if the marshaller is nil", func(t *testing.T) {
		t.Parallel()

		tx, err := process.GetTransactionHandler(
			0,
			0,
			hash,
			shardedDataCacherNotifier,
			storageService,
			nil,
			process.SearchMethodJustPeek)

		assert.Nil(t, tx)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
}

func TestGetTransactionHandlerShouldGetTransactionFromPool(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	txFromPool := &transaction.Transaction{Nonce: 1}

	storageService := &storageStubs.ChainStorerStub{}
	shardedDataCacherNotifier := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return txFromPool, true
				},
			}
		},
	}
	marshalizer := &mock.MarshalizerMock{}

	tx, err := process.GetTransactionHandler(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		storageService,
		marshalizer,
		process.SearchMethodJustPeek)

	assert.Nil(t, err)
	assert.Equal(t, tx, txFromPool)
}

func TestGetTransactionHandlerShouldGetTransactionFromStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	txFromStorage := &transaction.Transaction{
		Nonce: 1,
		Value: big.NewInt(0),
	}

	marshalizer := &mock.MarshalizerMock{}
	txMarsh, _ := marshalizer.Marshal(txFromStorage)
	storageService := &storageStubs.ChainStorerStub{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			if bytes.Equal(key, hash) {
				return txMarsh, nil
			}
			return nil, errors.New("error")
		},
	}
	shardedDataCacherNotifier := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}

	tx, err := process.GetTransactionHandler(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		storageService,
		marshalizer,
		process.SearchMethodJustPeek)

	assert.Nil(t, err)
	assert.Equal(t, tx, txFromStorage)
}

func TestGetTransactionHandlerFromPool_Errors(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	shardedDataCacherNotifier := testscommon.NewShardedDataStub()
	shardedDataCacherNotifier.ShardDataStoreCalled = func(cacheID string) storage.Cacher {
		return testscommon.NewCacherMock()
	}

	t.Run("nil sharded cache", func(t *testing.T) {
		tx, err := process.GetTransactionHandlerFromPool(
			0,
			0,
			hash,
			nil,
			process.SearchMethodJustPeek)

		assert.Nil(t, tx)
		assert.Equal(t, process.ErrNilShardedDataCacherNotifier, err)
	})
	t.Run("invalid method", func(t *testing.T) {
		tx, err := process.GetTransactionHandlerFromPool(
			0,
			0,
			hash,
			shardedDataCacherNotifier,
			166)

		assert.Nil(t, tx)
		assert.True(t, errors.Is(err, process.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "invalid value"))
		assert.True(t, strings.Contains(err.Error(), "166"))
	})
	t.Run("nil cache", func(t *testing.T) {
		shardedDataCacherNotifier.ShardDataStoreCalled = func(cacheID string) storage.Cacher {
			return nil
		}

		tx, err := process.GetTransactionHandlerFromPool(
			0,
			0,
			hash,
			shardedDataCacherNotifier,
			process.SearchMethodJustPeek)

		assert.Nil(t, tx)
		assert.Equal(t, process.ErrNilStorage, err)
	})
}

func TestGetTransactionHandlerFromPoolShouldErrTxNotFound(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	shardedDataCacherNotifier := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}

	tx, err := process.GetTransactionHandlerFromPool(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		process.SearchMethodJustPeek)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrTxNotFound, err)
}

func TestGetTransactionHandlerFromPoolShouldErrInvalidTxInPool(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	shardedDataCacherNotifier := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}

	tx, err := process.GetTransactionHandlerFromPool(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		process.SearchMethodJustPeek)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrInvalidTxInPool, err)
}

func TestGetTransactionHandlerFromPoolShouldWorkWithPeek(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	txFromPool := &transaction.Transaction{Nonce: 1}

	shardedDataCacherNotifier := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return txFromPool, true
				},
			}
		},
	}

	tx, err := process.GetTransactionHandlerFromPool(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		process.SearchMethodJustPeek)

	assert.Nil(t, err)
	assert.Equal(t, txFromPool, tx)
}

func TestGetTransactionHandlerFromPoolShouldWorkWithSearchFirst(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	txFromPool := &transaction.Transaction{Nonce: 1}

	shardedDataCacherNotifier := &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return txFromPool, true
		},
	}

	tx, err := process.GetTransactionHandlerFromPool(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		process.SearchMethodSearchFirst)

	assert.Nil(t, err)
	assert.Equal(t, txFromPool, tx)
}

func TestGetTransactionHandlerFromPoolShouldWorkWithPeekFallbackToSearchFirst(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	txFromPool := &transaction.Transaction{Nonce: 1}

	peekCalled := false
	shardedDataCacherNotifier := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					peekCalled = true
					return nil, false
				},
			}
		},
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return txFromPool, true
		},
	}

	tx, err := process.GetTransactionHandlerFromPool(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		process.SearchMethodPeekWithFallbackSearchFirst)

	assert.Nil(t, err)
	assert.Equal(t, txFromPool, tx)
	assert.True(t, peekCalled)
}

func TestGetTransactionHandlerFromStorageShouldErrNilStorage(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}

	tx, err := process.GetTransactionHandlerFromStorage(
		hash,
		nil,
		marshalizer)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetTransactionHandlerFromStorageShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	storageService := &storageStubs.ChainStorerStub{}

	tx, err := process.GetTransactionHandlerFromStorage(
		hash,
		storageService,
		nil)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetTransactionHandlerFromStorageShouldErrWhenTxIsNotFound(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	errExpected := errors.New("error")

	storageService := &storageStubs.ChainStorerStub{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return nil, errExpected
		},
	}
	marshalizer := &mock.MarshalizerMock{}

	tx, err := process.GetTransactionHandlerFromStorage(
		hash,
		storageService,
		marshalizer)

	assert.Nil(t, tx)
	assert.Equal(t, errExpected, err)
}

func TestGetTransactionHandlerFromStorageShouldErrWhenUnmarshalFail(t *testing.T) {
	t.Parallel()

	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &storageStubs.ChainStorerStub{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return nil, nil
		},
	}

	tx, err := process.GetTransactionHandlerFromStorage(
		hash,
		storageService,
		marshalizer)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
}

func TestGetTransactionHandlerFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("X")
	txFromPool := &transaction.Transaction{
		Nonce: 1,
		Value: big.NewInt(0),
	}

	marshalizer := &mock.MarshalizerMock{}
	txMarsh, _ := marshalizer.Marshal(txFromPool)
	storageService := &storageStubs.ChainStorerStub{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			if bytes.Equal(key, hash) {
				return txMarsh, nil
			}
			return nil, errors.New("error")
		},
	}

	tx, err := process.GetTransactionHandlerFromStorage(
		hash,
		storageService,
		marshalizer)

	assert.Nil(t, err)
	assert.Equal(t, txFromPool, tx)
}

func TestSortHeadersByNonceShouldWork(t *testing.T) {
	t.Parallel()

	headers := []data.HeaderHandler{
		&block.Header{Nonce: 3},
		&block.Header{Nonce: 2},
		&block.Header{Nonce: 1},
	}

	assert.Equal(t, uint64(3), headers[0].GetNonce())
	assert.Equal(t, uint64(2), headers[1].GetNonce())
	assert.Equal(t, uint64(1), headers[2].GetNonce())

	process.SortHeadersByNonce(headers)

	assert.Equal(t, uint64(1), headers[0].GetNonce())
	assert.Equal(t, uint64(2), headers[1].GetNonce())
	assert.Equal(t, uint64(3), headers[2].GetNonce())
}

func TestGetFinalCrossMiniBlockHashes(t *testing.T) {
	t.Parallel()

	hash1 := "hash1"
	hash2 := "hash2"

	mbh1 := block.MiniBlockHeader{
		SenderShardID: 1,
		Hash:          []byte(hash1),
	}
	mbhReserved1 := block.MiniBlockHeaderReserved{State: block.Proposed}
	mbh1.Reserved, _ = mbhReserved1.Marshal()

	mbh2 := block.MiniBlockHeader{
		SenderShardID: 2,
		Hash:          []byte(hash2),
	}
	mbhReserved2 := block.MiniBlockHeaderReserved{State: block.Final}
	mbh2.Reserved, _ = mbhReserved2.Marshal()

	header := &block.MetaBlock{
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbh1,
			mbh2,
		},
	}

	expectedHashes := map[string]uint32{
		hash2: 2,
	}

	hashes := process.GetFinalCrossMiniBlockHashes(header, 0)
	assert.Equal(t, expectedHashes, hashes)
}

func TestGetMiniBlockHeaderWithHash(t *testing.T) {
	t.Parallel()

	hash1, hash2 := "hash1", "hash2"

	t.Run("not equal hashes", func(t *testing.T) {
		t.Parallel()

		expectedMbh := &block.MiniBlockHeader{
			Hash: []byte(hash1),
		}
		header := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{*expectedMbh},
		}

		mbh := process.GetMiniBlockHeaderWithHash(header, []byte(hash2))
		assert.Nil(t, mbh)
	})

	t.Run("hashes matches", func(t *testing.T) {
		t.Parallel()

		expectedMbh := &block.MiniBlockHeader{
			Hash: []byte(hash1),
		}
		header := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{*expectedMbh},
		}

		mbh := process.GetMiniBlockHeaderWithHash(header, []byte(hash1))
		assert.Equal(t, expectedMbh, mbh)
	})
}

func Test_IsBuiltinFuncCallWithParam(t *testing.T) {
	txDataNoFunction := []byte("dummy data")
	targetFunction := "function"
	nonTargetFunction := "differentFunction"
	suffix := "@dummy@params"
	txDataWithFunc := []byte(targetFunction + suffix)
	txDataNonTargetFunc := []byte(nonTargetFunction + suffix)

	t.Run("no function", func(t *testing.T) {
		require.False(t, process.IsBuiltinFuncCallWithParam(txDataNoFunction, targetFunction))
	})
	t.Run("non target function", func(t *testing.T) {
		require.False(t, process.IsBuiltinFuncCallWithParam(txDataNonTargetFunc, targetFunction))
	})
	t.Run("target function", func(t *testing.T) {
		require.True(t, process.IsBuiltinFuncCallWithParam(txDataWithFunc, targetFunction))
	})
}

func Test_IsSetGuardianCall(t *testing.T) {
	t.Parallel()

	setGuardianTxData := []byte("SetGuardian@xxxxxxxx")
	t.Run("should return false for tx with other builtin function call or random data", func(t *testing.T) {
		require.False(t, process.IsSetGuardianCall([]byte(core.BuiltInFunctionClaimDeveloperRewards+"@...")))
		require.False(t, process.IsSetGuardianCall([]byte("some random data")))
	})
	t.Run("should return false for tx with setGuardian without params (no builtin function call)", func(t *testing.T) {
		require.False(t, process.IsSetGuardianCall([]byte("SetGuardian")))
	})
	t.Run("should return true for setGuardian call with invalid num of params", func(t *testing.T) {
		require.True(t, process.IsSetGuardianCall([]byte("SetGuardian@xxx@xxx@xxx")))
	})
	t.Run("should return true for setGuardian call with empty param", func(t *testing.T) {
		require.True(t, process.IsSetGuardianCall([]byte("SetGuardian@")))
	})
	t.Run("should return true for setGuardian call", func(t *testing.T) {
		require.True(t, process.IsSetGuardianCall(setGuardianTxData))
	})
}

func TestCheckIfIndexesAreOutOfBound(t *testing.T) {
	t.Parallel()

	txHashes := [][]byte{[]byte("txHash1"), []byte("txHash2"), []byte("txHash3")}
	miniBlock := &block.MiniBlock{TxHashes: txHashes}

	indexOfFirstTxToBeProcessed := int32(1)
	indexOfLastTxToBeProcessed := int32(0)
	err := process.CheckIfIndexesAreOutOfBound(indexOfFirstTxToBeProcessed, indexOfLastTxToBeProcessed, miniBlock)
	assert.True(t, errors.Is(err, process.ErrIndexIsOutOfBound))

	indexOfFirstTxToBeProcessed = int32(-1)
	indexOfLastTxToBeProcessed = int32(0)
	err = process.CheckIfIndexesAreOutOfBound(indexOfFirstTxToBeProcessed, indexOfLastTxToBeProcessed, miniBlock)
	assert.True(t, errors.Is(err, process.ErrIndexIsOutOfBound))

	indexOfFirstTxToBeProcessed = int32(0)
	indexOfLastTxToBeProcessed = int32(len(txHashes))
	err = process.CheckIfIndexesAreOutOfBound(indexOfFirstTxToBeProcessed, indexOfLastTxToBeProcessed, miniBlock)
	assert.True(t, errors.Is(err, process.ErrIndexIsOutOfBound))

	indexOfFirstTxToBeProcessed = int32(0)
	indexOfLastTxToBeProcessed = int32(len(txHashes) - 1)
	err = process.CheckIfIndexesAreOutOfBound(indexOfFirstTxToBeProcessed, indexOfLastTxToBeProcessed, miniBlock)
	assert.Nil(t, err)
}

func TestUnmarshalHeader(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	shardHeaderV1 := &block.Header{Nonce: 42, EpochStartMetaHash: []byte{0xaa, 0xbb}}
	shardHeaderV2 := &block.HeaderV2{Header: &block.Header{Nonce: 43, EpochStartMetaHash: []byte{0xaa, 0xbb}}}
	metaHeader := &block.MetaBlock{Nonce: 7, ValidatorStatsRootHash: []byte{0xcc, 0xdd}}

	shardHeaderV1Buffer, _ := marshalizer.Marshal(shardHeaderV1)
	shardHeaderV2Buffer, _ := marshalizer.Marshal(shardHeaderV2)
	metaHeaderBuffer, _ := marshalizer.Marshal(metaHeader)

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		header, err := process.UnmarshalHeader(1, marshalizer, shardHeaderV1Buffer)
		assert.Nil(t, err)
		assert.Equal(t, shardHeaderV1, header)

		header, err = process.UnmarshalHeader(1, marshalizer, shardHeaderV2Buffer)
		assert.Nil(t, err)
		assert.Equal(t, shardHeaderV2, header)

		header, err = process.UnmarshalHeader(core.MetachainShardId, marshalizer, metaHeaderBuffer)
		assert.Nil(t, err)
		assert.Equal(t, metaHeader, header)
	})

	t.Run("should err", func(t *testing.T) {
		t.Parallel()

		header, err := process.UnmarshalHeader(1, marshalizer, []byte{0xb, 0xa, 0xd})
		assert.NotNil(t, err)
		assert.Nil(t, header)

		header, err = process.UnmarshalHeader(core.MetachainShardId, marshalizer, []byte{0xb, 0xa, 0xd})
		assert.NotNil(t, err)
		assert.Nil(t, header)
	})
}

func TestShardedCacheSearchMethod_ToString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "just peek", process.SearchMethodJustPeek.ToString())
	assert.Equal(t, "search first", process.SearchMethodSearchFirst.ToString())
	assert.Equal(t, "peek with fallback to search first", process.SearchMethodPeekWithFallbackSearchFirst.ToString())
	str := process.ShardedCacheSearchMethod(166).ToString()
	assert.Equal(t, "unknown method 166", str)
}
