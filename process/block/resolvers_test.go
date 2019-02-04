package block_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- headerResolver

// NewHeaderResolver

func TestNewHeaderResolver_NilResolverShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := block.NewHeaderResolver(
		nil,
		&mock.TransientDataPoolStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilResolver, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilTransientPoolShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := block.NewHeaderResolver(
		&mock.ResolverStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilTransientPool, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilTransientHeadersPoolShouldErr(t *testing.T) {
	t.Parallel()

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, err := block.NewHeaderResolver(
		&mock.ResolverStub{},
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilTransientHeadersNoncesPoolShouldErr(t *testing.T) {
	t.Parallel()

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return nil
	}

	hdrRes, err := block.NewHeaderResolver(
		&mock.ResolverStub{},
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersStorageShouldErr(t *testing.T) {
	t.Parallel()

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, err := block.NewHeaderResolver(
		&mock.ResolverStub{},
		transientPool,
		nil,
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilHeadersStorage, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, err := block.NewHeaderResolver(
		&mock.ResolverStub{},
		transientPool,
		&mock.StorerStub{},
		nil,
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilNonceConverterShouldErr(t *testing.T) {
	t.Parallel()

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, err := block.NewHeaderResolver(
		&mock.ResolverStub{},
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilNonceConverter, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
		wasCalled = true
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, err := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.NotNil(t, hdrRes)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

// resolveHdrRequest

func TestHeaderResolver_ResolveHdrRequestNilValueShouldErr(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	buff, err := hdrRes.ResolveHdrRequest(process.RequestData{Type: process.NonceType, Value: nil})
	assert.Nil(t, buff)
	assert.Equal(t, process.ErrNilValue, err)

}

func TestHeaderResolver_ResolveHdrRequestUnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	buff, err := hdrRes.ResolveHdrRequest(process.RequestData{Type: 254, Value: make([]byte, 0)})
	assert.Nil(t, buff)
	assert.Equal(t, process.ErrResolveTypeUnknown, err)

}

func TestHeaderResolver_ResolveHdrRequestHashTypeFoundInHdrPoolShouldRetValue(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	requestedData := []byte("aaaa")
	resolvedData := []byte("bbbb")

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				return resolvedData, true
			}
			return nil, false
		}

		return headers
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	buff, err := hdrRes.ResolveHdrRequest(process.RequestData{Type: process.HashType, Value: requestedData})
	assert.Nil(t, err)

	recoveredResolved := make([]byte, 0)
	err = marshalizer.Unmarshal(&recoveredResolved, buff)
	assert.Equal(t, resolvedData, recoveredResolved)
}

func TestHeaderResolver_ResolveHdrRequestHashTypeFoundInHdrPoolMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	requestedData := []byte("aaaa")
	resolvedData := []byte("bbbb")

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				return resolvedData, true
			}
			return nil, false
		}

		return headers
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (i []byte, e error) {
		return nil, errors.New("MarshalizerMock generic error")
	}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	buff, err := hdrRes.ResolveHdrRequest(process.RequestData{Type: process.HashType, Value: requestedData})
	assert.Equal(t, "MarshalizerMock generic error", err.Error())
	assert.Nil(t, buff)
}

func TestHeaderResolver_ResolveHdrRequestRetFromStorageShouldRetVal(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	requestedData := []byte("aaaa")
	resolvedData := []byte("bbbb")

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			return resolvedData, nil
		}

		return nil, nil
	}

	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		store,
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	buff, _ := hdrRes.ResolveHdrRequest(process.RequestData{Type: process.HashType, Value: requestedData})
	assert.Equal(t, resolvedData, buff)
}

func TestHeaderResolver_ResolveHdrRequestRetFromStorageCheckRetError(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	requestedData := []byte("aaaa")
	resolvedData := []byte("bbbb")

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			return resolvedData, errors.New("just checking output error")
		}

		return nil, nil
	}

	marshalizer := &mock.MarshalizerMock{}
	marshalizer.Fail = true

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		store,
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	_, err := hdrRes.ResolveHdrRequest(process.RequestData{Type: process.HashType, Value: requestedData})
	assert.Equal(t, "just checking output error", err.Error())
}

func TestHeaderResolver_ResolveHdrRequestNonceTypeInvalidSliceShouldErr(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	buff, err := hdrRes.ResolveHdrRequest(process.RequestData{Type: process.NonceType, Value: []byte("aaa")})
	assert.Nil(t, buff)
	assert.Equal(t, process.ErrInvalidNonceByteSlice, err)
}

func TestHeaderResolver_ResolveHdrRequestNonceTypeNotFoundInHdrNoncePoolShouldRetNil(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		headersNonces := &mock.Uint64CacherStub{}
		headersNonces.GetCalled = func(u uint64) (i []byte, b bool) {
			return nil, false
		}

		return headersNonces
	}

	nonceConverter := mock.NewNonceHashConverterMock()

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
	)

	buff, err := hdrRes.ResolveHdrRequest(process.RequestData{
		Type:  process.NonceType,
		Value: nonceConverter.ToByteSlice(requestedNonce)})
	assert.Nil(t, buff)
	assert.Nil(t, err)

}

func TestHeaderResolver_ResolveHdrRequestNonceTypeFoundInHdrNoncePoolShouldRetFromPool(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	resolvedData := []byte("bbbb")

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, []byte("aaaa")) {
				return resolvedData, true
			}

			return nil, false
		}

		return headers
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		headersNonces := &mock.Uint64CacherStub{}
		headersNonces.GetCalled = func(u uint64) (i []byte, b bool) {
			if u == requestedNonce {
				return []byte("aaaa"), true
			}

			return nil, false
		}

		return headersNonces
	}

	nonceConverter := mock.NewNonceHashConverterMock()
	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		&mock.StorerStub{},
		marshalizer,
		nonceConverter,
	)

	buff, err := hdrRes.ResolveHdrRequest(process.RequestData{
		Type:  process.NonceType,
		Value: nonceConverter.ToByteSlice(requestedNonce)})
	assert.Nil(t, err)

	recoveredResolved := make([]byte, 0)
	err = marshalizer.Unmarshal(&recoveredResolved, buff)
	assert.Equal(t, resolvedData, recoveredResolved)
}

func TestHeaderResolver_ResolveHdrRequestNonceTypeFoundInHdrNoncePoolShouldRetFromStorage(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	resolvedData := []byte("bbbb")

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		headersNonces := &mock.Uint64CacherStub{}
		headersNonces.GetCalled = func(u uint64) (i []byte, b bool) {
			if u == requestedNonce {
				return []byte("aaaa"), true
			}

			return nil, false
		}

		return headersNonces
	}

	nonceConverter := mock.NewNonceHashConverterMock()
	marshalizer := &mock.MarshalizerMock{}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, []byte("aaaa")) {
			return resolvedData, nil
		}

		return nil, nil
	}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		store,
		marshalizer,
		nonceConverter,
	)

	buff, _ := hdrRes.ResolveHdrRequest(process.RequestData{
		Type:  process.NonceType,
		Value: nonceConverter.ToByteSlice(requestedNonce)})
	assert.Equal(t, resolvedData, buff)
}

func TestHeaderResolver_ResolveHdrRequestNonceTypeFoundInHdrNoncePoolCheckRetErr(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	resolvedData := []byte("bbbb")

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		headersNonces := &mock.Uint64CacherStub{}
		headersNonces.GetCalled = func(u uint64) (i []byte, b bool) {
			if u == requestedNonce {
				return []byte("aaaa"), true
			}

			return nil, false
		}

		return headersNonces
	}

	nonceConverter := mock.NewNonceHashConverterMock()
	marshalizer := &mock.MarshalizerMock{}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, []byte("aaaa")) {
			return resolvedData, errors.New("just checking output error")
		}

		return nil, nil
	}

	hdrRes, _ := block.NewHeaderResolver(
		topicResolver,
		transientPool,
		store,
		marshalizer,
		nonceConverter,
	)

	_, err := hdrRes.ResolveHdrRequest(process.RequestData{
		Type:  process.NonceType,
		Value: nonceConverter.ToByteSlice(requestedNonce)})
	assert.Equal(t, "just checking output error", err.Error())
}

// Requests

func TestHeaderResolver_RequestHdrFromHashShouldWork(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	requested := process.RequestData{}

	res.RequestDataCalled = func(rd process.RequestData) error {
		requested = rd
		return nil
	}

	buffRequested := []byte("aaaa")

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	nonceConverter := mock.NewNonceHashConverterMock()

	hdrRes, _ := block.NewHeaderResolver(
		res,
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
	)

	assert.Nil(t, hdrRes.RequestHeaderFromHash(buffRequested))
	assert.Equal(t, process.RequestData{
		Type:  process.HashType,
		Value: buffRequested,
	}, requested)
}

func TestHeaderResolver_RequestHdrFromNonceShouldWork(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	requested := process.RequestData{}

	res.RequestDataCalled = func(rd process.RequestData) error {
		requested = rd
		return nil
	}

	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	nonceConverter := mock.NewNonceHashConverterMock()

	hdrRes, _ := block.NewHeaderResolver(
		res,
		transientPool,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
	)

	buffToExpect := nonceConverter.ToByteSlice(67)

	assert.Nil(t, hdrRes.RequestHeaderFromNonce(67))
	assert.Equal(t, process.RequestData{
		Type:  process.NonceType,
		Value: buffToExpect,
	}, requested)
}

//------- genericBlockBodyResolver

// NewBlockBodyResolver

func TestNewGenericBlockBodyResolver_NilResolverShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := block.NewGenericBlockBodyResolver(
		nil,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilResolver, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockBodyPoolShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := block.NewGenericBlockBodyResolver(
		&mock.ResolverStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilBlockBodyPool, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockBodyStorageShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := block.NewGenericBlockBodyResolver(
		&mock.ResolverStub{},
		&mock.CacherStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilBlockBodyStorage, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockMArshalizerShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := block.NewGenericBlockBodyResolver(
		&mock.ResolverStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
		wasCalled = true
	}

	gbbRes, err := block.NewGenericBlockBodyResolver(
		res,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, gbbRes)
	assert.True(t, wasCalled)
}

// resolveBlockBodyRequest

func TestGenericBlockBodyResolver_ResolveBlockBodyRequestWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	gbbRes, _ := block.NewGenericBlockBodyResolver(
		topicResolver,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	buff, err := gbbRes.ResolveBlockBodyRequest(process.RequestData{Type: process.NonceType, Value: make([]byte, 0)})
	assert.Nil(t, buff)
	assert.Equal(t, process.ErrResolveNotHashType, err)

}

func TestGenericBlockBodyResolver_ResolveBlockBodyRequestNilValueShouldErr(t *testing.T) {
	t.Parallel()

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	gbbRes, _ := block.NewGenericBlockBodyResolver(
		topicResolver,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	buff, err := gbbRes.ResolveBlockBodyRequest(process.RequestData{Type: process.HashType, Value: nil})
	assert.Nil(t, buff)
	assert.Equal(t, process.ErrNilValue, err)

}

func TestGenericBlockBodyResolver_ResolveBlockBodyRequestFoundInPoolShouldRetVal(t *testing.T) {
	t.Parallel()

	requestedBuff := []byte("aaa")
	resolvedBuff := []byte("bbb")

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	cache := &mock.CacherStub{}
	cache.GetCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, requestedBuff) {
			return resolvedBuff, true
		}

		return nil, false
	}

	marshalizer := &mock.MarshalizerMock{}

	gbbRes, _ := block.NewGenericBlockBodyResolver(
		topicResolver,
		cache,
		&mock.StorerStub{},
		marshalizer,
	)

	buff, err := gbbRes.ResolveBlockBodyRequest(process.RequestData{
		Type:  process.HashType,
		Value: requestedBuff})

	buffExpected, _ := marshalizer.Marshal(resolvedBuff)

	assert.Nil(t, err)
	assert.Equal(t, buffExpected, buff)

}

func TestGenericBlockBodyResolver_ResolveBlockBodyRequestFoundInPoolMarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	requestedBuff := []byte("aaa")
	resolvedBuff := []byte("bbb")

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	cache := &mock.CacherStub{}
	cache.GetCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, requestedBuff) {
			return resolvedBuff, true
		}

		return nil, false
	}

	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (i []byte, e error) {
		return nil, errors.New("MarshalizerMock generic error")
	}

	gbbRes, _ := block.NewGenericBlockBodyResolver(
		topicResolver,
		cache,
		&mock.StorerStub{},
		marshalizer,
	)

	buff, err := gbbRes.ResolveBlockBodyRequest(process.RequestData{
		Type:  process.HashType,
		Value: requestedBuff})

	assert.Nil(t, buff)
	assert.Equal(t, "MarshalizerMock generic error", err.Error())

}

func TestGenericBlockBodyResolver_ResolveBlockBodyRequestNotFoundInPoolShouldRetFromStorage(t *testing.T) {
	t.Parallel()

	requestedBuff := []byte("aaa")
	resolvedBuff := []byte("bbb")

	topicResolver := &mock.ResolverStub{}
	topicResolver.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
	}

	cache := &mock.CacherStub{}
	cache.GetCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		return resolvedBuff, errors.New("just checking output error")
	}

	marshalizer := &mock.MarshalizerMock{}

	gbbRes, _ := block.NewGenericBlockBodyResolver(
		topicResolver,
		cache,
		store,
		marshalizer,
	)

	buff, err := gbbRes.ResolveBlockBodyRequest(process.RequestData{
		Type:  process.HashType,
		Value: requestedBuff})

	assert.Equal(t, resolvedBuff, buff)
	assert.Equal(t, "just checking output error", err.Error())

}

// Requests

func TestBlockBodyResolver_RequestBlockBodyFromHashShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	buffRequested := []byte("aaaa")

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
		wasCalled = true
	}

	requested := process.RequestData{}

	res.RequestDataCalled = func(rd process.RequestData) error {
		requested = rd
		return nil
	}

	gbbRes, err := block.NewGenericBlockBodyResolver(
		res,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, gbbRes)
	assert.True(t, wasCalled)

	assert.Nil(t, gbbRes.RequestBlockBodyFromHash(buffRequested))
	assert.Equal(t, process.RequestData{
		Type:  process.HashType,
		Value: buffRequested,
	}, requested)
}
