package resolvers_test

//TODO(jls) fix this
//
//import (
//	"bytes"
//	"errors"
//	"testing"
//
//	"github.com/ElrondNetwork/elrond-go/dataRetriever"
//	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
//	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
//	"github.com/ElrondNetwork/elrond-go/p2p"
//	"github.com/stretchr/testify/assert"
//)
//
////------- NewHeaderResolver
//
//func TestNewHeaderResolver_NilSenderResolverShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, err := resolvers.NewHeaderResolver(
//		nil,
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
//	assert.Nil(t, hdrRes)
//}
//
//func TestNewHeaderResolver_NilHeadersPoolShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, err := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		nil,
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	assert.Equal(t, dataRetriever.ErrNilHeadersDataPool, err)
//	assert.Nil(t, hdrRes)
//}
//
//func TestNewHeaderResolver_NilHeadersNoncesPoolShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, err := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		&mock.CacherStub{},
//		nil,
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	assert.Equal(t, dataRetriever.ErrNilHeadersNoncesDataPool, err)
//	assert.Nil(t, hdrRes)
//}
//
//func TestNewHeaderResolver_NilHeadersStorageShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, err := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		nil,
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	assert.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
//	assert.Nil(t, hdrRes)
//}
//
//func TestNewHeaderResolver_NilNonceConverterShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, err := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		nil,
//	)
//
//	assert.Equal(t, dataRetriever.ErrNilNonceConverter, err)
//	assert.Nil(t, hdrRes)
//}
//
//func TestNewHeaderResolver_OkValsShouldWork(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, err := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	assert.NotNil(t, hdrRes)
//	assert.Nil(t, err)
//}
//
////------- ProcessReceivedMessage
//
//func TestHeaderResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, nil))
//	assert.Equal(t, dataRetriever.ErrNilValue, err)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestUnknownTypeShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(254, make([]byte, 0)))
//	assert.Equal(t, dataRetriever.ErrResolveTypeUnknown, err)
//
//}
//
//func TestHeaderResolver_ValidateRequestHashTypeFoundInHdrPoolShouldSearchAndSend(t *testing.T) {
//	t.Parallel()
//
//	requestedData := []byte("aaaa")
//
//	searchWasCalled := false
//	sendWasCalled := false
//
//	headers := &mock.CacherStub{}
//
//	headers.PeekCalled = func(key []byte) (value interface{}, ok bool) {
//		if bytes.Equal(requestedData, key) {
//			searchWasCalled = true
//			return make([]byte, 0), true
//		}
//		return nil, false
//	}
//
//	marshalizer := &mock.MarshalizerMock{}
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendCalled: func(buff []byte, peer p2p.PeerID) error {
//				sendWasCalled = true
//				return nil
//			},
//		},
//		headers,
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		marshalizer,
//		mock.NewNonceHashConverterMock(),
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData))
//	assert.Nil(t, err)
//	assert.True(t, searchWasCalled)
//	assert.True(t, sendWasCalled)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestHashTypeFoundInHdrPoolMarshalizerFailsShouldErr(t *testing.T) {
//	t.Parallel()
//
//	requestedData := []byte("aaaa")
//	resolvedData := []byte("bbbb")
//
//	errExpected := errors.New("MarshalizerMock generic error")
//
//	headers := &mock.CacherStub{}
//
//	headers.PeekCalled = func(key []byte) (value interface{}, ok bool) {
//		if bytes.Equal(requestedData, key) {
//			return resolvedData, true
//		}
//		return nil, false
//	}
//
//	marshalizerMock := &mock.MarshalizerMock{}
//	marshalizerStub := &mock.MarshalizerStub{
//		MarshalCalled: func(obj interface{}) (i []byte, e error) {
//			return nil, errExpected
//		},
//		UnmarshalCalled: func(obj interface{}, buff []byte) error {
//			return marshalizerMock.Unmarshal(obj, buff)
//		},
//	}
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendCalled: func(buff []byte, peer p2p.PeerID) error {
//				return nil
//			},
//		},
//		headers,
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		marshalizerStub,
//		mock.NewNonceHashConverterMock(),
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData))
//	assert.Equal(t, errExpected, err)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestRetFromStorageShouldRetValAndSend(t *testing.T) {
//	t.Parallel()
//
//	requestedData := []byte("aaaa")
//
//	headers := &mock.CacherStub{}
//
//	headers.PeekCalled = func(key []byte) (value interface{}, ok bool) {
//		return nil, false
//	}
//
//	wasGotFromStorage := false
//	wasSent := false
//
//	store := &mock.StorerStub{}
//	store.GetCalled = func(key []byte) (i []byte, e error) {
//		if bytes.Equal(key, requestedData) {
//			wasGotFromStorage = true
//			return make([]byte, 0), nil
//		}
//
//		return nil, nil
//	}
//
//	marshalizer := &mock.MarshalizerMock{}
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendCalled: func(buff []byte, peer p2p.PeerID) error {
//				wasSent = true
//				return nil
//			},
//		},
//		headers,
//		&mock.Uint64CacherStub{},
//		store,
//		marshalizer,
//		mock.NewNonceHashConverterMock(),
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData))
//	assert.Nil(t, err)
//	assert.True(t, wasGotFromStorage)
//	assert.True(t, wasSent)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeInvalidSliceShouldErr(t *testing.T) {
//	t.Parallel()
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{},
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		mock.NewNonceHashConverterMock(),
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("aaa")))
//	assert.Equal(t, dataRetriever.ErrInvalidNonceByteSlice, err)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeNotFoundInHdrNoncePoolShouldRetNilAndNotSend(t *testing.T) {
//	t.Parallel()
//
//	requestedNonce := uint64(67)
//
//	headersNonces := &mock.Uint64CacherStub{}
//	headersNonces.GetCalled = func(u uint64) (i interface{}, b bool) {
//		return nil, false
//	}
//
//	nonceConverter := mock.NewNonceHashConverterMock()
//
//	wasSent := false
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendCalled: func(buff []byte, peer p2p.PeerID) error {
//				wasSent = true
//				return nil
//			},
//		},
//		&mock.CacherStub{},
//		headersNonces,
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		nonceConverter,
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
//		dataRetriever.NonceType,
//		nonceConverter.ToByteSlice(requestedNonce)))
//	assert.Nil(t, err)
//	assert.False(t, wasSent)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolShouldRetFromPoolAndSend(t *testing.T) {
//	t.Parallel()
//
//	requestedNonce := uint64(67)
//
//	wasResolved := false
//	wasSent := false
//
//	headers := &mock.CacherStub{}
//	headers.PeekCalled = func(key []byte) (value interface{}, ok bool) {
//		if bytes.Equal(key, []byte("aaaa")) {
//			wasResolved = true
//			return make([]byte, 0), true
//		}
//
//		return nil, false
//	}
//
//	headersNonces := &mock.Uint64CacherStub{}
//	headersNonces.GetCalled = func(u uint64) (i interface{}, b bool) {
//		if u == requestedNonce {
//			return []byte("aaaa"), true
//		}
//
//		return nil, false
//	}
//
//	nonceConverter := mock.NewNonceHashConverterMock()
//	marshalizer := &mock.MarshalizerMock{}
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendCalled: func(buff []byte, peer p2p.PeerID) error {
//				wasSent = true
//				return nil
//			},
//		},
//		headers,
//		headersNonces,
//		&mock.StorerStub{},
//		marshalizer,
//		nonceConverter,
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
//		dataRetriever.NonceType,
//		nonceConverter.ToByteSlice(requestedNonce)))
//
//	assert.Nil(t, err)
//	assert.True(t, wasResolved)
//	assert.True(t, wasSent)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolShouldRetFromStorageAndSend(t *testing.T) {
//	t.Parallel()
//
//	requestedNonce := uint64(67)
//
//	wasResolved := false
//	wasSend := false
//
//	headers := &mock.CacherStub{}
//	headers.PeekCalled = func(key []byte) (value interface{}, ok bool) {
//		return nil, false
//	}
//
//	headersNonces := &mock.Uint64CacherStub{}
//	headersNonces.GetCalled = func(u uint64) (i interface{}, b bool) {
//		if u == requestedNonce {
//			return []byte("aaaa"), true
//		}
//
//		return nil, false
//	}
//
//	nonceConverter := mock.NewNonceHashConverterMock()
//	marshalizer := &mock.MarshalizerMock{}
//
//	store := &mock.StorerStub{}
//	store.GetCalled = func(key []byte) (i []byte, e error) {
//		if bytes.Equal(key, []byte("aaaa")) {
//			wasResolved = true
//			return make([]byte, 0), nil
//		}
//
//		return nil, nil
//	}
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendCalled: func(buff []byte, peer p2p.PeerID) error {
//				wasSend = true
//				return nil
//			},
//		},
//		headers,
//		headersNonces,
//		store,
//		marshalizer,
//		nonceConverter,
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
//		dataRetriever.NonceType,
//		nonceConverter.ToByteSlice(requestedNonce)))
//
//	assert.Nil(t, err)
//	assert.True(t, wasResolved)
//	assert.True(t, wasSend)
//}
//
//func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolCheckRetErr(t *testing.T) {
//	t.Parallel()
//
//	requestedNonce := uint64(67)
//
//	errExpected := errors.New("expected error")
//
//	headers := &mock.CacherStub{}
//	headers.PeekCalled = func(key []byte) (value interface{}, ok bool) {
//		return nil, false
//	}
//
//	headersNonces := &mock.Uint64CacherStub{}
//	headersNonces.GetCalled = func(u uint64) (i interface{}, b bool) {
//		if u == requestedNonce {
//			return []byte("aaaa"), true
//		}
//
//		return nil, false
//	}
//
//	nonceConverter := mock.NewNonceHashConverterMock()
//	marshalizer := &mock.MarshalizerMock{}
//
//	store := &mock.StorerStub{}
//	store.GetCalled = func(key []byte) (i []byte, e error) {
//		if bytes.Equal(key, []byte("aaaa")) {
//			return nil, errExpected
//		}
//
//		return nil, nil
//	}
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendCalled: func(buff []byte, peer p2p.PeerID) error {
//				return nil
//			},
//		},
//		headers,
//		headersNonces,
//		store,
//		marshalizer,
//		nonceConverter,
//	)
//
//	err := hdrRes.ProcessReceivedMessage(createRequestMsg(
//		dataRetriever.NonceType,
//		nonceConverter.ToByteSlice(requestedNonce)))
//
//	assert.Equal(t, errExpected, err)
//}
//
////------- Requests
//
//func TestHeaderResolver_RequestDataFromNonceShouldWork(t *testing.T) {
//	t.Parallel()
//
//	nonceRequested := uint64(67)
//	wasRequested := false
//
//	nonceConverter := mock.NewNonceHashConverterMock()
//
//	buffToExpect := nonceConverter.ToByteSlice(nonceRequested)
//
//	hdrRes, _ := resolvers.NewHeaderResolver(
//		&mock.TopicResolverSenderStub{
//			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
//				if bytes.Equal(rd.Value, buffToExpect) {
//					wasRequested = true
//				}
//				return nil
//			},
//		},
//		&mock.CacherStub{},
//		&mock.Uint64CacherStub{},
//		&mock.StorerStub{},
//		&mock.MarshalizerMock{},
//		nonceConverter,
//	)
//
//	assert.Nil(t, hdrRes.RequestDataFromNonce(nonceRequested))
//	assert.True(t, wasRequested)
//}
