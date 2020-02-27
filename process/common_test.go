package process_test

import (
	"bytes"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestEmptyChannelShouldWorkOnBufferedChannel(t *testing.T) {
	ch := make(chan bool, 10)

	assert.Equal(t, 0, len(ch))
	readsCnt := process.EmptyChannel(ch)
	assert.Equal(t, 0, len(ch))
	assert.Equal(t, 0, readsCnt)

	ch <- true
	ch <- true
	ch <- true

	assert.Equal(t, 3, len(ch))
	readsCnt = process.EmptyChannel(ch)
	assert.Equal(t, 0, len(ch))
	assert.Equal(t, 3, readsCnt)
}

func TestEmptyChannelShouldWorkOnNotBufferedChannel(t *testing.T) {
	ch := make(chan bool)

	assert.Equal(t, 0, len(ch))
	readsCnt := int32(process.EmptyChannel(ch))
	assert.Equal(t, 0, len(ch))
	assert.Equal(t, int32(0), readsCnt)

	wg := sync.WaitGroup{}
	wgChanWasWritten := sync.WaitGroup{}
	numConcurrentWrites := 50
	wg.Add(numConcurrentWrites)
	wgChanWasWritten.Add(numConcurrentWrites)
	for i := 0; i < numConcurrentWrites; i++ {
		go func() {
			wg.Done()
			time.Sleep(time.Millisecond)
			ch <- true
			wgChanWasWritten.Done()
		}()
	}

	// wait for go routines to start
	wg.Wait()

	go func() {
		for readsCnt < int32(numConcurrentWrites) {
			atomic.AddInt32(&readsCnt, int32(process.EmptyChannel(ch)))
		}
	}()

	// wait for go routines to finish
	wgChanWasWritten.Wait()

	assert.Equal(t, 0, len(ch))
	assert.Equal(t, int32(numConcurrentWrites), atomic.LoadInt32(&readsCnt))
}

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

	cacher := &mock.HeadersCacherStub{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetShardHeader(hash, cacher, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetShardHeader(hash, cacher, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetShardHeaderShouldGetHeaderFromPool(t *testing.T) {
	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return hdr, nil
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
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
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

func TestGetMetaHeaderShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetMetaHeader(hash, nil, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderShouldErrNilMarshalizer(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetMetaHeader(hash, cacher, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMetaHeaderShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetMetaHeader(hash, cacher, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMetaHeaderShouldGetHeaderFromPool(t *testing.T) {
	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 1}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return hdr, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, _ := process.GetMetaHeader(hash, cacher, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderShouldGetHeaderFromStorage(t *testing.T) {
	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 1}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
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

	header, _ := process.GetMetaHeader(hash, cacher, marshalizer, storageService)
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

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}

	header, err := process.GetShardHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromPoolShouldErrWrongTypeAssertion(t *testing.T) {
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
	hash := []byte("X")

	header, err := process.GetMetaHeaderFromPool(hash, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderFromPoolShouldErrMissingHeader(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}

	header, err := process.GetMetaHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMetaHeaderFromPoolShouldErrWrongTypeAssertion(t *testing.T) {
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

func TestGetMetaHeaderFromStorageShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	storageService := &mock.ChainStorerMock{}

	header, err := process.GetMetaHeaderFromStorage(hash, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMetaHeaderFromStorageShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMetaHeaderFromStorageShouldErrNilHeadersStorage(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilHeadersStorage, err)
}

func TestGetMetaHeaderFromStorageShouldErrMissingHeader(t *testing.T) {
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

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMetaHeaderFromStorageShouldErrUnmarshalWithoutSuccess(t *testing.T) {
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

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
}

func TestGetMetaHeaderFromStorageShouldWork(t *testing.T) {
	hash := []byte("X")

	hdr := &block.MetaBlock{}
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

	header, err := process.GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	assert.Nil(t, err)
	assert.Equal(t, hdr, header)
}

func TestGetMarshalizedHeaderFromStorageShouldErrNilMarshalizer(t *testing.T) {
	hash := []byte("X")

	storageService := &mock.ChainStorerMock{}

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, nil, storageService)
	assert.Nil(t, headerMarsh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMarshalizedHeaderFromStorageShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, nil)
	assert.Nil(t, headerMarsh)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMarshalizedHeaderFromStorageShouldErrNilHeadersStorage(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	assert.Nil(t, headerMarsh)
	assert.Equal(t, process.ErrNilHeadersStorage, err)
}

func TestGetMarshalizedHeaderFromStorageShouldErrMissingHeader(t *testing.T) {
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

	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	assert.Nil(t, headerMarsh)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMarshalizedHeaderFromStorageShouldWork(t *testing.T) {
	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
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

	hdrMarsh, err := marshalizer.Marshal(hdr)
	headerMarsh, err := process.GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	assert.Equal(t, hdrMarsh, headerMarsh)
	assert.Nil(t, err)
}

func TestGetShardHeaderWithNonceShouldErrNilCacher(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}
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
	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}
	storageService := &mock.ChainStorerMock{}
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
	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

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
	storageService := &mock.ChainStorerMock{}
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
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}
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
	nonce := uint64(1)

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}
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
	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}
	storageService := &mock.ChainStorerMock{}
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
	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

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
	hash := []byte("X")
	nonce := uint64(1)

	hdr := &block.MetaBlock{Nonce: nonce}
	cacher := &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}
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
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}
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
	nonce := uint64(1)
	shardId := uint32(0)

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, nil)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenShardIdHashMapIsNil(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenLoadFromShardIdHashMapFails(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromPoolWithNonceShouldErrMissingHeader(t *testing.T) {
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
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromPoolWithNonceShouldErrWrongTypeAssertion(t *testing.T) {
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
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromPoolWithNonceShouldWork(t *testing.T) {
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
	nonce := uint64(1)

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, nil)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenShardIdHashMapIsNil(t *testing.T) {
	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrMissingHashForHeaderNonceWhenLoadFromShardIdHashMapFails(t *testing.T) {
	hash := []byte("X")
	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{}

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrMissingHeader(t *testing.T) {
	hash := []byte("X")
	nonce := uint64(1)

	cacher := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
			return nil, errors.New("err")
		},
	}

	header, hash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher)
	assert.Nil(t, header)
	assert.Nil(t, hash)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMetaHeaderFromPoolWithNonceShouldErrWrongTypeAssertion(t *testing.T) {
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
	nonce := uint64(1)
	shardId := uint32(0)

	storageService := &mock.ChainStorerMock{}
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
	nonce := uint64(1)
	shardId := uint32(0)

	storageService := &mock.ChainStorerMock{}
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

func TestGetShardHeaderFromStorageWithNonceShouldErrNilHeadersStorage(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)

	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
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
	assert.Equal(t, process.ErrNilHeadersStorage, err)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrMissingHashForHeaderNonce(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)

	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}
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
	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					return nil, errors.New("error")
				},
			}
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
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetShardHeaderFromStorageWithNonceShouldErrUnmarshalWithoutSuccess(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return nil, nil
					}
					return nil, errors.New("error")
				},
			}
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

func TestGetShardHeaderFromStorageWithNonceShouldWork(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	nonceToByte := []byte("1")
	hdr := &block.Header{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}
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

func TestGetHeaderFromStorageWithNonceShouldWorkForShard(t *testing.T) {
	nonce := uint64(1)
	shardId := uint32(0)
	hash := []byte("X")
	nonceToByte := []byte("1")
	hdr := &block.Header{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}
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
		shardId,
		storageService,
		uint64Converter,
		marshalizer)

	assert.Nil(t, err)
	assert.Equal(t, hash, headerHash)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrNilStorage(t *testing.T) {
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
	nonce := uint64(1)

	storageService := &mock.ChainStorerMock{}
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
	nonce := uint64(1)

	storageService := &mock.ChainStorerMock{}
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

func TestGetMetaHeaderFromStorageWithNonceShouldErrNilHeadersStorage(t *testing.T) {
	nonce := uint64(1)

	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
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
	assert.Equal(t, process.ErrNilHeadersStorage, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrMissingHashForHeaderNonce(t *testing.T) {
	nonce := uint64(1)

	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("error")
				},
			}
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
	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					return nil, errors.New("error")
				},
			}
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
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMetaHeaderFromStorageWithNonceShouldErrUnmarshalWithoutSuccess(t *testing.T) {
	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return nil, nil
					}
					return nil, errors.New("error")
				},
			}
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
	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	hdr := &block.MetaBlock{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}
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
	nonce := uint64(1)
	hash := []byte("X")
	nonceToByte := []byte("1")
	hdr := &block.MetaBlock{Nonce: nonce}
	marshalizer := &mock.MarshalizerMock{}
	marshHdr, _ := marshalizer.Marshal(hdr)
	storageService := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByte) {
						return hash, nil
					}
					if bytes.Equal(key, hash) {
						return marshHdr, nil
					}
					return nil, errors.New("error")
				},
			}
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

func TestGetTransactionHandlerShouldErrNilShardedDataCacherNotifier(t *testing.T) {
	hash := []byte("X")

	storageService := &mock.ChainStorerMock{}
	marshalizer := &mock.MarshalizerMock{}

	tx, err := process.GetTransactionHandler(
		0,
		0,
		hash,
		nil,
		storageService,
		marshalizer,
		false)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilShardedDataCacherNotifier, err)
}

func TestGetTransactionHandlerShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	shardedDataCacherNotifier := &mock.ShardedDataStub{}
	marshalizer := &mock.MarshalizerMock{}

	tx, err := process.GetTransactionHandler(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		nil,
		marshalizer,
		false)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetTransactionHandlerShouldErrNilMarshalizer(t *testing.T) {
	hash := []byte("X")

	storageService := &mock.ChainStorerMock{}
	shardedDataCacherNotifier := &mock.ShardedDataStub{}

	tx, err := process.GetTransactionHandler(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		storageService,
		nil,
		false)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetTransactionHandlerShouldGetTransactionFromPool(t *testing.T) {
	hash := []byte("X")
	txFromPool := &transaction.Transaction{Nonce: 1}

	storageService := &mock.ChainStorerMock{}
	shardedDataCacherNotifier := &mock.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &mock.CacherStub{
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
		false)

	assert.Nil(t, err)
	assert.Equal(t, tx, txFromPool)
}

func TestGetTransactionHandlerShouldGetTransactionFromStorage(t *testing.T) {
	hash := []byte("X")
	txFromStorage := &transaction.Transaction{
		Nonce: 1,
		Value: big.NewInt(0),
	}

	marshalizer := &mock.MarshalizerMock{}
	txMarsh, _ := marshalizer.Marshal(txFromStorage)
	storageService := &mock.ChainStorerMock{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			if bytes.Equal(key, hash) {
				return txMarsh, nil
			}
			return nil, errors.New("error")
		},
	}
	shardedDataCacherNotifier := &mock.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &mock.CacherStub{
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
		false)

	assert.Nil(t, err)
	assert.Equal(t, tx, txFromStorage)
}

func TestGetTransactionHandlerFromPoolShouldErrNilShardedDataCacherNotifier(t *testing.T) {
	hash := []byte("X")

	tx, err := process.GetTransactionHandlerFromPool(
		0,
		0,
		hash,
		nil,
		false)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilShardedDataCacherNotifier, err)
}

func TestGetTransactionHandlerFromPoolShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	shardedDataCacherNotifier := &mock.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return nil
		},
	}

	tx, err := process.GetTransactionHandlerFromPool(
		0,
		0,
		hash,
		shardedDataCacherNotifier,
		false)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetTransactionHandlerFromPoolShouldErrTxNotFound(t *testing.T) {
	hash := []byte("X")

	shardedDataCacherNotifier := &mock.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &mock.CacherStub{
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
		false)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrTxNotFound, err)
}

func TestGetTransactionHandlerFromPoolShouldErrInvalidTxInPool(t *testing.T) {
	hash := []byte("X")

	shardedDataCacherNotifier := &mock.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &mock.CacherStub{
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
		false)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrInvalidTxInPool, err)
}

func TestGetTransactionHandlerFromPoolShouldWork(t *testing.T) {
	hash := []byte("X")
	txFromPool := &transaction.Transaction{Nonce: 1}

	shardedDataCacherNotifier := &mock.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &mock.CacherStub{
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
		false)

	assert.Nil(t, err)
	assert.Equal(t, txFromPool, tx)
}

func TestGetTransactionHandlerFromStorageShouldErrNilStorage(t *testing.T) {
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
	hash := []byte("X")

	storageService := &mock.ChainStorerMock{}

	tx, err := process.GetTransactionHandlerFromStorage(
		hash,
		storageService,
		nil)

	assert.Nil(t, tx)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetTransactionHandlerFromStorageShouldErrWhenTxIsNotFound(t *testing.T) {
	hash := []byte("X")
	errExpected := errors.New("error")

	storageService := &mock.ChainStorerMock{
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
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{
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
	hash := []byte("X")
	txFromPool := &transaction.Transaction{
		Nonce: 1,
		Value: big.NewInt(0),
	}

	marshalizer := &mock.MarshalizerMock{}
	txMarsh, _ := marshalizer.Marshal(txFromPool)
	storageService := &mock.ChainStorerMock{
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
