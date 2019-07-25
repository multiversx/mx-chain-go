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

func TestEmptyChannelShouldWork(t *testing.T) {
	ch := make(chan bool, 10)

	ch <- true
	ch <- true
	ch <- true

	assert.Equal(t, 3, len(ch))
	process.EmptyChannel(ch)
	assert.Equal(t, 0, len(ch))
}

func TestGetShardHeaderShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}

	header, err := process.GetShardHeader(hash, nil, uint64SyncMapCacherStub, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetShardHeaderShouldErrNilUint64SyncMapCacher(t *testing.T) {
	hash := []byte("X")

	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}
	cacher := &mock.CacherStub{}

	header, err := process.GetShardHeader(hash, cacher, nil, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilUint64SyncMapCacher, err)
}

func TestGetShardHeaderShouldErrNilMarshalizer(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetShardHeader(hash, cacher, uint64SyncMapCacherStub, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetShardHeaderShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetShardHeader(hash, cacher, uint64SyncMapCacherStub, marshalizer, nil)
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
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, _ := process.GetShardHeader(hash, cacher, uint64SyncMapCacherStub, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetShardHeaderShouldGetHeaderFromStorage(t *testing.T) {
	hash := []byte("X")

	hdr := &block.Header{Nonce: 1}
	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		PutCalled: func(key []byte, value interface{}) (evicted bool) {
			return true
		},
	}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{
		MergeCalled: func(nonce uint64, src dataRetriever.ShardIdHashMap) {
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

	header, _ := process.GetShardHeader(hash, cacher, uint64SyncMapCacherStub, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetMetaHeader(hash, nil, uint64SyncMapCacherStub, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderShouldErrNilUint64SyncMapCacher(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetMetaHeader(hash, cacher, nil, marshalizer, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilUint64SyncMapCacher, err)
}

func TestGetMetaHeaderShouldErrNilMarshalizer(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}
	storageService := &mock.ChainStorerMock{}

	header, err := process.GetMetaHeader(hash, cacher, uint64SyncMapCacherStub, nil, storageService)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestGetMetaHeaderShouldErrNilStorage(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}
	marshalizer := &mock.MarshalizerMock{}

	header, err := process.GetMetaHeader(hash, cacher, uint64SyncMapCacherStub, marshalizer, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestGetMetaHeaderShouldGetHeaderFromPool(t *testing.T) {
	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 1}
	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return hdr, true
		},
	}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{}
	marshalizer := &mock.MarshalizerMock{}
	storageService := &mock.ChainStorerMock{}

	header, _ := process.GetMetaHeader(hash, cacher, uint64SyncMapCacherStub, marshalizer, storageService)
	assert.Equal(t, hdr, header)
}

func TestGetMetaHeaderShouldGetHeaderFromStorage(t *testing.T) {
	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 1}
	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		PutCalled: func(key []byte, value interface{}) (evicted bool) {
			return true
		},
	}
	uint64SyncMapCacherStub := &mock.Uint64SyncMapCacherStub{
		MergeCalled: func(nonce uint64, src dataRetriever.ShardIdHashMap) {
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

	header, _ := process.GetMetaHeader(hash, cacher, uint64SyncMapCacherStub, marshalizer, storageService)
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

func TestGetMetaHeaderFromPoolShouldErrNilCacher(t *testing.T) {
	hash := []byte("X")

	header, err := process.GetMetaHeaderFromPool(hash, nil)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestGetMetaHeaderFromPoolShouldErrMissingHeader(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}

	header, err := process.GetMetaHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestGetMetaHeaderFromPoolShouldErrWrongTypeAssertion(t *testing.T) {
	hash := []byte("X")

	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return &block.Header{}, true
		},
	}

	header, err := process.GetMetaHeaderFromPool(hash, cacher)
	assert.Nil(t, header)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestGetMetaHeaderFromPoolShouldWork(t *testing.T) {
	hash := []byte("X")

	hdr := &block.MetaBlock{Nonce: 10}
	cacher := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return hdr, true
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
