package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewHeaderResolver

func TestNewHeaderResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		nil,
		&mock.PoolsHolderStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilResolverSender, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilPoolsHolderShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilPoolsHolder, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersPoolShouldErr(t *testing.T) {
	t.Parallel()

	pools := createDataPool()
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		pools,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersNoncesPoolShouldErr(t *testing.T) {
	t.Parallel()

	pools := createDataPool()
	pools.HeadersNoncesCalled = func() data.Uint64Cacher {
		return nil
	}

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		pools,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersStorageShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		createDataPool(),
		nil,
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilHeadersStorage, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		createDataPool(),
		&mock.StorerStub{},
		nil,
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilNonceConverterShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		createDataPool(),
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilNonceConverter, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		createDataPool(),
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.NotNil(t, hdrRes)
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestHeaderResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		createDataPool(),
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(process.NonceType, nil))
	assert.Equal(t, process.ErrNilValue, err)
}

func TestHeaderResolver_ProcessReceivedMessageRequestUnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		createDataPool(),
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(254, make([]byte, 0)))
	assert.Equal(t, process.ErrResolveTypeUnknown, err)

}

func TestHeaderResolver_ValidateRequestHashTypeFoundInHdrPoolShouldSearchAndSend(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")

	searchWasCalled := false
	sendWasCalled := false

	pools := createDataPool()
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				searchWasCalled = true
				return make([]byte, 0), true
			}
			return nil, false
		}

		return headers
	}

	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		pools,
		&mock.StorerStub{},
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
}

func TestHeaderResolver_ProcessReceivedMessageRequestHashTypeFoundInHdrPoolMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	resolvedData := []byte("bbbb")

	errExpected := errors.New("MarshalizerMock generic error")

	pools := createDataPool()
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				return resolvedData, true
			}
			return nil, false
		}

		return headers
	}

	marshalizerMock := &mock.MarshalizerMock{}
	marshalizerStub := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			return nil, errExpected
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return marshalizerMock.Unmarshal(obj, buff)
		},
	}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				return nil
			},
		},
		pools,
		&mock.StorerStub{},
		marshalizerStub,
		mock.NewNonceHashConverterMock(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Equal(t, errExpected, err)
}

func TestHeaderResolver_ProcessReceivedMessageRequestRetFromStorageShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")

	pools := createDataPool()
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}

	wasGotFromStorage := false
	wasSent := false

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			wasGotFromStorage = true
			return make([]byte, 0), nil
		}

		return nil, nil
	}

	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		pools,
		store,
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Nil(t, err)
	assert.True(t, wasGotFromStorage)
	assert.True(t, wasSent)
}

func TestHeaderResolver_ProcessReceivedMessageRequestRetFromStorageCheckRetError(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")

	pools := createDataPool()
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}

	errExpected := errors.New("expected error")

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			return nil, errExpected
		}

		return nil, nil
	}

	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		pools,
		store,
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Equal(t, errExpected, err)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeInvalidSliceShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		createDataPool(),
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(process.NonceType, []byte("aaa")))
	assert.Equal(t, process.ErrInvalidNonceByteSlice, err)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeNotFoundInHdrNoncePoolShouldRetNilAndNotSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)

	pools := createDataPool()
	pools.HeadersNoncesCalled = func() data.Uint64Cacher {
		headersNonces := &mock.Uint64CacherStub{}
		headersNonces.GetCalled = func(u uint64) (i []byte, b bool) {
			return nil, false
		}

		return headersNonces
	}

	nonceConverter := mock.NewNonceHashConverterMock()

	wasSent := false

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		pools,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
		process.NonceType,
		nonceConverter.ToByteSlice(requestedNonce)))
	assert.Nil(t, err)
	assert.False(t, wasSent)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolShouldRetFromPoolAndSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)

	wasResolved := false
	wasSent := false

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, []byte("aaaa")) {
				wasResolved = true
				return make([]byte, 0), true
			}

			return nil, false
		}

		return headers
	}
	pools.HeadersNoncesCalled = func() data.Uint64Cacher {
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

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		pools,
		&mock.StorerStub{},
		marshalizer,
		nonceConverter,
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
		process.NonceType,
		nonceConverter.ToByteSlice(requestedNonce)))

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolShouldRetFromStorageAndSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)

	wasResolved := false
	wasSend := false

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}
	pools.HeadersNoncesCalled = func() data.Uint64Cacher {
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
			wasResolved = true
			return make([]byte, 0), nil
		}

		return nil, nil
	}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSend = true
				return nil
			},
		},
		pools,
		store,
		marshalizer,
		nonceConverter,
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
		process.NonceType,
		nonceConverter.ToByteSlice(requestedNonce)))

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolCheckRetErr(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)

	errExpected := errors.New("expected error")

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() data.ShardedDataCacherNotifier {
		headers := &mock.ShardedDataStub{}

		headers.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return headers
	}
	pools.HeadersNoncesCalled = func() data.Uint64Cacher {
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
			return nil, errExpected
		}

		return nil, nil
	}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				return nil
			},
		},
		pools,
		store,
		marshalizer,
		nonceConverter,
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
		process.NonceType,
		nonceConverter.ToByteSlice(requestedNonce)))

	assert.Equal(t, errExpected, err)
}

//------- Requests

func TestHeaderResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	buffRequested := []byte("aaaa")

	wasRequested := false

	nonceConverter := mock.NewNonceHashConverterMock()

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *process.RequestData) error {
				if bytes.Equal(rd.Value, buffRequested) {
					wasRequested = true
				}

				return nil
			},
		},
		createDataPool(),
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
	)

	assert.Nil(t, hdrRes.RequestDataFromHash(buffRequested))
	assert.True(t, wasRequested)
}

func TestHeaderResolver_RequestDataFromNonceShouldWork(t *testing.T) {
	t.Parallel()

	nonceRequested := uint64(67)
	wasRequested := false

	nonceConverter := mock.NewNonceHashConverterMock()

	buffToExpect := nonceConverter.ToByteSlice(nonceRequested)

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *process.RequestData) error {
				if bytes.Equal(rd.Value, buffToExpect) {
					wasRequested = true
				}
				return nil
			},
		},
		createDataPool(),
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
	)

	assert.Nil(t, hdrRes.RequestDataFromNonce(nonceRequested))
	assert.True(t, wasRequested)
}
