package process_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/estimator"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
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

	testGetMetaHeader(t, &block.MetaBlock{Nonce: 1})
	testGetMetaHeader(t, &block.MetaBlockV3{Nonce: 2})
}

func testGetMetaHeader(t *testing.T, hdr data.MetaHeaderHandler) {
	hash := []byte("X")
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

	testGetMetaHeaderFromPool(t, &block.MetaBlock{Nonce: 10})
	testGetMetaHeaderFromPool(t, &block.MetaBlockV3{Nonce: 11})
}

func testGetMetaHeaderFromPool(t *testing.T, hdr data.MetaHeaderHandler) {
	hash := []byte("X")
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

	testGetMetaHeaderFromStorage(t, &block.MetaBlock{Nonce: 1})
	testGetMetaHeaderFromStorage(t, &block.MetaBlockV3{Nonce: 2, LastExecutionResult: &block.MetaExecutionResultInfo{}})
}

func testGetMetaHeaderFromStorage(t *testing.T, hdr data.MetaHeaderHandler) {
	hash := []byte("X")

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

	testGetMetaHeaderWithNonceShouldGetHeaderFromStorage(t, &block.MetaBlock{Nonce: 1})
	testGetMetaHeaderWithNonceShouldGetHeaderFromStorage(t, &block.MetaBlockV3{Nonce: 2})
}

func testGetMetaHeaderWithNonceShouldGetHeaderFromStorage(t *testing.T, hdr data.MetaHeaderHandler) {
	hash := []byte("X")
	nonceToByte := []byte("1")

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
			if n == hdr.GetNonce() {
				return nonceToByte
			}

			return nil
		},
	}

	header, _, _ := process.GetMetaHeaderWithNonce(
		hdr.GetNonce(),
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

	testGetMetaHeaderFromPoolWithNonce(t, &block.MetaBlock{Nonce: 1})
	testGetMetaHeaderFromPoolWithNonce(t, &block.MetaBlockV3{Nonce: 2})
}

func testGetMetaHeaderFromPoolWithNonce(t *testing.T, hdr data.MetaHeaderHandler) {
	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}

	header, headerHash, err := process.GetMetaHeaderFromPoolWithNonce(hdr.GetNonce(), cacher)
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

	testGetMetaHeaderFromStorageWithNonce(t, &block.MetaBlock{Nonce: 1})
	testGetMetaHeaderFromStorageWithNonce(t, &block.MetaBlockV3{Nonce: 2, LastExecutionResult: &block.MetaExecutionResultInfo{}})
}

func testGetMetaHeaderFromStorageWithNonce(t *testing.T, hdr data.MetaHeaderHandler) {
	hash := []byte("X")
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
			if n == hdr.GetNonce() {
				return nonceToByte
			}

			return nil
		},
	}

	header, headerHash, err := process.GetMetaHeaderFromStorageWithNonce(
		hdr.GetNonce(),
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
			return &cache.CacherStub{
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
			return &cache.CacherStub{
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
		return cache.NewCacherMock()
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
			return &cache.CacherStub{
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
			return &cache.CacherStub{
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
			return &cache.CacherStub{
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
			return &cache.CacherStub{
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
	shardHeaderV3 := &block.HeaderV3{Nonce: 44, LastExecutionResult: &block.ExecutionResultInfo{NotarizedInRound: 100}}
	metaHeader := &block.MetaBlock{Nonce: 7, ValidatorStatsRootHash: []byte{0xcc, 0xdd}}
	metaHeaderV3 := &block.MetaBlockV3{Nonce: 8, LastExecutionResult: &block.MetaExecutionResultInfo{NotarizedInRound: 100}}

	shardHeaderV1Buffer, _ := marshalizer.Marshal(shardHeaderV1)
	shardHeaderV2Buffer, _ := marshalizer.Marshal(shardHeaderV2)
	shardHeaderV3Buffer, _ := marshalizer.Marshal(shardHeaderV3)
	metaHeaderBuffer, _ := marshalizer.Marshal(metaHeader)
	metaHeaderV3Buffer, _ := marshalizer.Marshal(metaHeaderV3)

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		header, err := process.UnmarshalHeader(1, marshalizer, shardHeaderV1Buffer)
		assert.Nil(t, err)
		assert.Equal(t, shardHeaderV1, header)

		header, err = process.UnmarshalHeader(1, marshalizer, shardHeaderV2Buffer)
		assert.Nil(t, err)
		assert.Equal(t, shardHeaderV2, header)

		header, err = process.UnmarshalHeader(1, marshalizer, shardHeaderV3Buffer)
		assert.Nil(t, err)
		assert.Equal(t, shardHeaderV3, header)

		header, err = process.UnmarshalHeader(core.MetachainShardId, marshalizer, metaHeaderBuffer)
		assert.Nil(t, err)
		assert.Equal(t, metaHeader, header)

		header, err = process.UnmarshalHeader(core.MetachainShardId, marshalizer, metaHeaderV3Buffer)
		assert.Nil(t, err)
		assert.Equal(t, metaHeaderV3, header)
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

func Test_SetBaseExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("nil execution manager should error", func(t *testing.T) {
		t.Parallel()
		err := process.SetBaseExecutionResult(nil, &testscommon.ChainHandlerStub{})
		require.Equal(t, process.ErrNilExecutionManager, err)
	})

	t.Run("nil chain handler should error", func(t *testing.T) {
		t.Parallel()
		err := process.SetBaseExecutionResult(&processMocks.ExecutionManagerMock{}, nil)
		require.Equal(t, process.ErrNilBlockChain, err)
	})

	t.Run("execution manager error should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("execution manager error")
		executionManager := &processMocks.ExecutionManagerMock{
			SetLastNotarizedResultCalled: func(data.BaseExecutionResultHandler) error {
				return expectedErr
			},
		}

		executionResultsInfo := &block.ExecutionResultInfo{
			NotarizedInRound: 100,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash"),
				HeaderNonce: 100,
				HeaderRound: 100,
				RootHash:    []byte("root hash"),
			},
		}

		header := &block.HeaderV3{
			Round:               100,
			LastExecutionResult: executionResultsInfo,
		}
		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
		}

		err := process.SetBaseExecutionResult(executionManager, chainHandler)
		require.Equal(t, expectedErr, err)
	})

	t.Run("chain handler returning nil should not set", func(t *testing.T) {
		t.Parallel()

		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{
			SetLastNotarizedResultCalled: func(data.BaseExecutionResultHandler) error {
				require.Fail(t, "should not be called")
				return nil
			},
		}

		err := process.SetBaseExecutionResult(executionManager, chainHandler)
		require.Nil(t, err)
	})

	t.Run("chain handler returning header of different version should not set", func(t *testing.T) {
		t.Parallel()

		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{
			SetLastNotarizedResultCalled: func(data.BaseExecutionResultHandler) error {
				require.Fail(t, "should not be called")
				return nil
			},
		}

		err := process.SetBaseExecutionResult(executionManager, chainHandler)
		require.Nil(t, err)
	})
	t.Run("chain handler returning header without execution result should not set", func(t *testing.T) {
		t.Parallel()

		header := &block.HeaderV3{}
		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{
			SetLastNotarizedResultCalled: func(data.BaseExecutionResultHandler) error {
				require.Fail(t, "should not be called")
				return nil
			},
		}

		err := process.SetBaseExecutionResult(executionManager, chainHandler)
		require.Equal(t, process.ErrNilLastExecutionResultHandler, err)
	})

	t.Run("chain handler returning header with execution result but nil based execution result should not set", func(t *testing.T) {
		t.Parallel()

		header := &block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{},
		}
		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{
			SetLastNotarizedResultCalled: func(data.BaseExecutionResultHandler) error {
				require.Fail(t, "should not be called")
				return nil
			},
		}

		err := process.SetBaseExecutionResult(executionManager, chainHandler)
		require.Equal(t, process.ErrNilBaseExecutionResult, err)
	})

	t.Run("ok with shard header", func(t *testing.T) {
		t.Parallel()

		executionResultsInfo := &block.ExecutionResultInfo{
			NotarizedInRound: 101,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash"),
				HeaderNonce: 100,
				HeaderRound: 100,
				RootHash:    []byte("root hash"),
			},
		}

		header := &block.HeaderV3{
			Round:               101,
			LastExecutionResult: executionResultsInfo,
		}

		called := false
		executionManager := &processMocks.ExecutionManagerMock{
			SetLastNotarizedResultCalled: func(result data.BaseExecutionResultHandler) error {
				require.Equal(t, executionResultsInfo.ExecutionResult, result)
				called = true
				return nil
			},
		}
		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
		}

		err := process.SetBaseExecutionResult(executionManager, chainHandler)
		require.Nil(t, err)
		require.True(t, called)
	})

	t.Run("ok with meta header", func(t *testing.T) {
		t.Parallel()

		executionResultsInfo := &block.MetaExecutionResultInfo{
			NotarizedInRound: 101,
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("hash"),
					HeaderNonce: 100,
					HeaderRound: 100,
					RootHash:    []byte("root hash"),
				},
			},
		}

		header := &block.MetaBlockV3{
			Round:               101,
			LastExecutionResult: executionResultsInfo,
		}

		called := false
		executionManager := &processMocks.ExecutionManagerMock{
			SetLastNotarizedResultCalled: func(result data.BaseExecutionResultHandler) error {
				require.Equal(t, executionResultsInfo.ExecutionResult, result)
				called = true
				return nil
			},
		}
		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
		}

		err := process.SetBaseExecutionResult(executionManager, chainHandler)
		require.Nil(t, err)
		require.True(t, called)
	})
}

func TestTransactionCoordinator_SeparateBody(t *testing.T) {
	t.Parallel()

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.TxBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.TxBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.TxBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.InvalidBlock}) // invalid blocks will go into the TxBlock bucket
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.PeerBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.PeerBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.RewardsBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.RewardsBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.RewardsBlock})

	separated := process.SeparateBodyByType(body)
	require.Equal(t, 4, len(separated))
	require.Equal(t, 4, len(separated[block.TxBlock].MiniBlocks))
	require.Equal(t, 4, len(separated[block.SmartContractResultBlock].MiniBlocks))
	require.Equal(t, 2, len(separated[block.PeerBlock].MiniBlocks))
	require.Equal(t, 3, len(separated[block.RewardsBlock].MiniBlocks))
}

func Test_CreateLastExecutionResultInfoFromExecutionResult(t *testing.T) {
	t.Parallel()

	notarizedInRound := uint64(1)
	t.Run("nil executionResult", func(t *testing.T) {
		t.Parallel()

		lastExecutionResultInfo, err := process.CreateLastExecutionResultInfoFromExecutionResult(notarizedInRound, nil, 0)
		require.Equal(t, process.ErrNilExecutionResultHandler, err)
		require.Nil(t, lastExecutionResultInfo)
	})
	t.Run("invalid shard executionResult type", func(t *testing.T) {
		t.Parallel()

		executionResult := createDummyMetaExecutionResult() // This is a valid type, but we will use it incorrectly
		lastExecutionResultInfo, err := process.CreateLastExecutionResultInfoFromExecutionResult(notarizedInRound, executionResult, 0)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Nil(t, lastExecutionResultInfo)
	})
	t.Run("valid executionResult for shard", func(t *testing.T) {
		t.Parallel()

		executionResult := createDummyShardExecutionResult()
		expectedLastExecutionResult := &block.ExecutionResultInfo{
			NotarizedInRound: notarizedInRound,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  executionResult.GetHeaderHash(),
				HeaderNonce: executionResult.GetHeaderNonce(),
				HeaderRound: executionResult.GetHeaderRound(),
				RootHash:    executionResult.GetRootHash(),
				GasUsed:     executionResult.GetGasUsed(),
			},
		}

		lastExecutionResultHandler, err := process.CreateLastExecutionResultInfoFromExecutionResult(notarizedInRound, executionResult, 0)
		lastExecutionResultInfo := lastExecutionResultHandler.(*block.ExecutionResultInfo)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResultInfo)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResultInfo)
	})
	t.Run("invalid metaChain executionResult", func(t *testing.T) {
		t.Parallel()

		executionResult := createDummyShardExecutionResult()
		lastExecutionResultHandler, err := process.CreateLastExecutionResultInfoFromExecutionResult(notarizedInRound, executionResult, core.MetachainShardId)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Nil(t, lastExecutionResultHandler)
	})
	t.Run("valid executionResult for metaChain", func(t *testing.T) {
		t.Parallel()

		executionResult := createDummyMetaExecutionResult()
		expectedLastExecutionResult := &block.MetaExecutionResultInfo{
			NotarizedInRound: notarizedInRound,
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  executionResult.GetHeaderHash(),
					HeaderNonce: executionResult.GetHeaderNonce(),
					HeaderRound: executionResult.GetHeaderRound(),
					RootHash:    executionResult.GetRootHash(),
					GasUsed:     executionResult.GetGasUsed(),
				},
				ValidatorStatsRootHash: executionResult.GetValidatorStatsRootHash(),
				AccumulatedFeesInEpoch: executionResult.GetAccumulatedFeesInEpoch(),
				DevFeesInEpoch:         executionResult.GetDevFeesInEpoch(),
			},
		}

		lastExecutionResultHandler, err := process.CreateLastExecutionResultInfoFromExecutionResult(notarizedInRound, executionResult, core.MetachainShardId)
		lastExecutionResultInfo, ok := lastExecutionResultHandler.(*block.MetaExecutionResultInfo)
		require.True(t, ok)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResultInfo)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResultInfo)
	})
}

func Test_GetPrevBlockLastExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("nil blockchain", func(t *testing.T) {
		t.Parallel()

		_, err := process.GetPrevBlockLastExecutionResult(nil)
		require.Equal(t, process.ErrNilBlockChain, err)
	})
	t.Run("nil current block header", func(t *testing.T) {
		t.Parallel()

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}
		_, err := process.GetPrevBlockLastExecutionResult(blockchain)
		require.Equal(t, process.ErrNilHeaderHandler, err)
	})
	t.Run("valid prev header - shard header v3", func(t *testing.T) {
		t.Parallel()

		prevHeader := createShardHeaderV3WithExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prevHeaderHash")
			},
		}

		lastExecutionResult, err := process.GetPrevBlockLastExecutionResult(blockchain)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		require.Equal(t, prevHeader.LastExecutionResult, lastExecutionResult)
	})
	t.Run("valid prev header - shard header v2", func(t *testing.T) {
		t.Parallel()

		prevHeader := createDummyPrevShardHeaderV2()
		prevHeaderHash := []byte("prevHeaderHash")
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}

		lastExecutionResult, err := process.GetPrevBlockLastExecutionResult(blockchain)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		expectedLastExecutionResult, _ := common.CreateLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResult)
	})
	t.Run("valid prev header - meta header v3", func(t *testing.T) {
		t.Parallel()

		prevHeader := createMetaHeaderV3WithExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prevHeaderHash")
			},
		}

		lastExecutionResult, err := process.GetPrevBlockLastExecutionResult(blockchain)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		require.Equal(t, prevHeader.LastExecutionResult, lastExecutionResult)
	})
	t.Run("valid prev header - meta header", func(t *testing.T) {
		t.Parallel()

		prevHeader := createDummyPrevMetaHeader()
		prevHeaderHash := []byte("prevHeaderHash")
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}

		lastExecutionResult, err := process.GetPrevBlockLastExecutionResult(blockchain)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		expectedLastExecutionResult, _ := common.CreateLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResult)
	})
}

func Test_CreateDataForInclusionEstimation(t *testing.T) {
	t.Parallel()

	t.Run("nil last exec result handler", func(t *testing.T) {
		t.Parallel()

		lastExecResForInclusion, err := process.CreateDataForInclusionEstimation(nil)
		require.Equal(t, process.ErrNilLastExecutionResultHandler, err)
		require.Nil(t, lastExecResForInclusion)
	})
	t.Run("valid last execution result for shard", func(t *testing.T) {
		t.Parallel()

		executionResult := createLastExecutionResultShard()
		expectedLastExecResultForInclusion := &estimator.LastExecutionResultForInclusion{
			NotarizedInRound: executionResult.GetNotarizedInRound(),
			ProposedInRound:  executionResult.GetExecutionResult().GetHeaderRound(),
		}
		lastExecResForInclusion, err := process.CreateDataForInclusionEstimation(executionResult)
		require.NoError(t, err)
		require.NotNil(t, lastExecResForInclusion)
		require.Equal(t, expectedLastExecResultForInclusion, lastExecResForInclusion)
	})
	t.Run("valid last execution result for meta", func(t *testing.T) {
		t.Parallel()

		executionResult := createLastExecutionResultMeta()
		expectedLastExecResultForInclusion := &estimator.LastExecutionResultForInclusion{
			NotarizedInRound: executionResult.GetNotarizedInRound(),
			ProposedInRound:  executionResult.GetExecutionResult().GetHeaderRound(),
		}
		lastExecResForInclusion, err := process.CreateDataForInclusionEstimation(executionResult)
		require.NoError(t, err)
		require.NotNil(t, lastExecResForInclusion)
		require.Equal(t, expectedLastExecResultForInclusion, lastExecResForInclusion)
	})
	t.Run("invalid last execution result", func(t *testing.T) {
		t.Parallel()

		executionResult := &block.ExecutionResult{}
		lastExecResForInclusion, err := process.CreateDataForInclusionEstimation(executionResult)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Nil(t, lastExecResForInclusion)
	})
}

func Test_IsNotExecutableTransactionError(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		require.False(t, process.IsNotExecutableTransactionError(nil))
	})
	t.Run("lower nonce", func(t *testing.T) {
		t.Parallel()
		require.True(t, process.IsNotExecutableTransactionError(process.ErrLowerNonceInTransaction))
	})
	t.Run("higher nonce", func(t *testing.T) {
		t.Parallel()
		wrappedErr := fmt.Errorf("wrapping: %w", process.ErrHigherNonceInTransaction)
		require.True(t, process.IsNotExecutableTransactionError(wrappedErr))
	})
	t.Run("insufficient fee", func(t *testing.T) {
		t.Parallel()
		require.True(t, process.IsNotExecutableTransactionError(process.ErrInsufficientFee))
	})
	t.Run("transaction is not executable", func(t *testing.T) {
		t.Parallel()
		require.True(t, process.IsNotExecutableTransactionError(process.ErrTransactionNotExecutable))
	})
	t.Run("different error", func(t *testing.T) {
		t.Parallel()
		require.False(t, process.IsNotExecutableTransactionError(errors.New("some other error")))
	})
}

func createDummyShardExecutionResult() *block.ExecutionResult {
	return &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("headerHash"),
			HeaderNonce: 1,
			HeaderRound: 2,
			GasUsed:     10000000,
			RootHash:    []byte("rootHash"),
		},
		ReceiptsHash:     []byte("receiptsHash"),
		MiniBlockHeaders: nil,
		DeveloperFees:    big.NewInt(100),
		AccumulatedFees:  big.NewInt(200),
		ExecutedTxCount:  100,
	}
}

func createDummyMetaExecutionResult() *block.MetaExecutionResult {
	return &block.MetaExecutionResult{
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("headerHash"),
				HeaderNonce: 1,
				HeaderRound: 2,
				GasUsed:     20000000,
				RootHash:    []byte("rootHash"),
			},
			ValidatorStatsRootHash: []byte("validatorStatsRootHash"),
			AccumulatedFeesInEpoch: big.NewInt(300),
			DevFeesInEpoch:         big.NewInt(400),
		},
		ReceiptsHash:    []byte("receiptsHash"),
		DeveloperFees:   big.NewInt(500),
		AccumulatedFees: big.NewInt(600),
		ExecutedTxCount: 200,
	}
}

func createDummyPrevShardHeaderV2() *block.HeaderV2 {
	return &block.HeaderV2{
		Header: &block.Header{
			Nonce:    1,
			Round:    2,
			RootHash: []byte("prevRootHash"),
		},
	}
}
func createDummyInvalidMetaHeader() data.HeaderHandler {
	return &block.HeaderV2{
		Header: &block.Header{
			Nonce:    1,
			Round:    2,
			RootHash: []byte("prevRootHash"),
			ShardID:  core.MetachainShardId,
		},
	}
}

func createLastExecutionResultShard() *block.ExecutionResultInfo {
	return &block.ExecutionResultInfo{
		NotarizedInRound: 3,
		ExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("headerHash"),
			HeaderNonce: 1,
			HeaderRound: 2,
			RootHash:    []byte("rootHash"),
		},
	}
}

func createLastExecutionResultMeta() *block.MetaExecutionResultInfo {
	return &block.MetaExecutionResultInfo{
		NotarizedInRound: 3,
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("headerHash"),
				HeaderNonce: 1,
				HeaderRound: 2,
				RootHash:    []byte("rootHash"),
			},
			ValidatorStatsRootHash: []byte("validatorStatsRootHash"),
			AccumulatedFeesInEpoch: big.NewInt(300),
			DevFeesInEpoch:         big.NewInt(400),
		},
	}
}

func createShardHeaderV3WithExecutionResults() *block.HeaderV3 {
	executionResult := createDummyShardExecutionResult()
	lastExecutionResult := createLastExecutionResultShard()
	return &block.HeaderV3{
		PrevHash:            []byte("prevHash"),
		Nonce:               2,
		Round:               3,
		ShardID:             0,
		ExecutionResults:    []*block.ExecutionResult{executionResult},
		LastExecutionResult: lastExecutionResult,
	}
}

func createMetaHeaderV3WithExecutionResults() *block.MetaBlockV3 {
	executionResults := createDummyMetaExecutionResult()
	lastExecutionResult := createLastExecutionResultMeta()
	return &block.MetaBlockV3{
		Nonce:               3,
		Round:               4,
		LastExecutionResult: lastExecutionResult,
		ExecutionResults:    []*block.MetaExecutionResult{executionResults},
	}
}

func createDummyPrevMetaHeader() *block.MetaBlock {
	return &block.MetaBlock{
		Nonce:                  4,
		Round:                  5,
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("prevValidatorStatsRootHash"),
		DevFeesInEpoch:         big.NewInt(300),
		AccumulatedFeesInEpoch: big.NewInt(400),
	}
}
