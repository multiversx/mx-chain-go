package storageResolvers

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/assert"
)

func createMockSliceResolverArg() ArgSliceResolver {
	return ArgSliceResolver{
		Messenger:                &mock.MessageHandlerStub{},
		ResponseTopicName:        "",
		Storage:                  genericmocks.NewStorerMock("Storage", 0),
		DataPacker:               &mock.DataPackerStub{},
		Marshalizer:              &mock.MarshalizerMock{},
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess),
	}
}

func TestNewSliceResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceResolverArg()
	arg.Messenger = nil
	sr, err := NewSliceResolver(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewSliceResolver_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceResolverArg()
	arg.Storage = nil
	sr, err := NewSliceResolver(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewSliceResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceResolverArg()
	arg.Marshalizer = nil
	sr, err := NewSliceResolver(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewSliceResolver_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceResolverArg()
	arg.DataPacker = nil
	sr, err := NewSliceResolver(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewSliceResolver_NilManualEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceResolverArg()
	arg.ManualEpochStartNotifier = nil
	sr, err := NewSliceResolver(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilManualEpochStartNotifier, err)
}

func TestNewSliceResolver_NilGracefullyCloseChanShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceResolverArg()
	arg.ChanGracefullyClose = nil
	sr, err := NewSliceResolver(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilGracefullyCloseChannel, err)
}

func TestNewSliceResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockSliceResolverArg()
	sr, err := NewSliceResolver(arg)

	assert.False(t, check.IfNil(sr))
	assert.Nil(t, err)
}

func TestSliceResolver_RequestDataFromHashNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	sendWasCalled := false
	arg := createMockSliceResolverArg()
	arg.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendWasCalled = true
			return nil
		},
	}
	arg.ChanGracefullyClose = make(chan endProcess.ArgEndProcess, 1)
	sr, _ := NewSliceResolver(arg)

	err := sr.RequestDataFromHash([]byte("hash"), 0)

	assert.Equal(t, expectedErr, err)
	assert.False(t, sendWasCalled)

	time.Sleep(time.Second)

	select {
	case argClose := <-arg.ChanGracefullyClose:
		assert.Equal(t, core.ImportComplete, argClose.Reason)
	default:
		assert.Fail(t, "did not wrote on end chan")
	}
}

func TestSliceResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	sendWasCalled := false
	arg := createMockSliceResolverArg()
	arg.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendWasCalled = true
			return nil
		},
	}
	sr, _ := NewSliceResolver(arg)

	err := sr.RequestDataFromHash([]byte("hash"), 0)

	assert.Nil(t, err)
	assert.True(t, sendWasCalled)
}

func TestSliceResolver_RequestDataFromHashesShouldWork(t *testing.T) {
	t.Parallel()

	numSendCalled := 0
	numGetCalled := 0
	arg := createMockSliceResolverArg()
	arg.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			numGetCalled++
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			numSendCalled++
			return nil
		},
	}
	sr, _ := NewSliceResolver(arg)

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	err := sr.RequestDataFromHashArray(hashes, 0)

	assert.Nil(t, err)
	assert.Equal(t, len(hashes), numSendCalled)
	assert.Equal(t, len(hashes), numGetCalled)
}

func TestSliceResolver_GetErroredShouldReturnErr(t *testing.T) {
	t.Parallel()

	numSendCalled := 0
	numGetCalled := 0
	expectedErr := errors.New("expected err")
	arg := createMockSliceResolverArg()
	arg.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			numGetCalled++
			if numGetCalled == 1 {
				return nil, expectedErr
			}

			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			numSendCalled++
			return nil
		},
	}
	arg.ChanGracefullyClose = make(chan endProcess.ArgEndProcess, 1)
	sr, _ := NewSliceResolver(arg)

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	err := sr.RequestDataFromHashArray(hashes, 0)

	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, len(hashes)-1, numSendCalled)
	assert.Equal(t, len(hashes), numGetCalled)

	time.Sleep(time.Second)

	select {
	case argClose := <-arg.ChanGracefullyClose:
		assert.Equal(t, core.ImportComplete, argClose.Reason)
	default:
		assert.Fail(t, "did not wrote on end chan")
	}
}

func TestSliceResolver_SendErroredShouldReturnErr(t *testing.T) {
	t.Parallel()

	numSendCalled := 0
	numGetCalled := 0
	expectedErr := errors.New("expected err")
	arg := createMockSliceResolverArg()
	arg.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			numGetCalled++
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			numSendCalled++
			if numSendCalled == 1 {
				return expectedErr
			}
			return nil
		},
	}
	sr, _ := NewSliceResolver(arg)

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	err := sr.RequestDataFromHashArray(hashes, 0)

	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, 1, numSendCalled)
	assert.Equal(t, len(hashes), numGetCalled)
}
