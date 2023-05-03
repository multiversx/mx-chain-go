package storagerequesters

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected err")

func createMockSliceRequesterArg() ArgSliceRequester {
	return ArgSliceRequester{
		Messenger:                &mock.MessageHandlerStub{},
		ResponseTopicName:        "",
		Storage:                  genericMocks.NewStorerMock(),
		DataPacker:               &mock.DataPackerStub{},
		Marshalizer:              &mock.MarshalizerMock{},
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess),
	}
}

func TestNewSliceRequester_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	arg.Messenger = nil
	sr, err := NewSliceRequester(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewSliceRequester_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	arg.Storage = nil
	sr, err := NewSliceRequester(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewSliceRequester_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	arg.Marshalizer = nil
	sr, err := NewSliceRequester(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewSliceRequester_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	arg.DataPacker = nil
	sr, err := NewSliceRequester(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewSliceRequester_NilManualEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	arg.ManualEpochStartNotifier = nil
	sr, err := NewSliceRequester(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilManualEpochStartNotifier, err)
}

func TestNewSliceRequester_NilGracefullyCloseChanShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	arg.ChanGracefullyClose = nil
	sr, err := NewSliceRequester(arg)

	assert.True(t, check.IfNil(sr))
	assert.Equal(t, dataRetriever.ErrNilGracefullyCloseChannel, err)
}

func TestNewSliceRequester_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	sr, err := NewSliceRequester(arg)

	assert.False(t, check.IfNil(sr))
	assert.Nil(t, err)
}

func TestSliceRequester_RequestDataFromHashNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	sendWasCalled := false
	arg := createMockSliceRequesterArg()
	arg.Storage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	arg.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendWasCalled = true
			return nil
		},
	}
	arg.ChanGracefullyClose = make(chan endProcess.ArgEndProcess, 1)
	sr, _ := NewSliceRequester(arg)

	err := sr.RequestDataFromHash([]byte("hash"), 0)

	assert.Equal(t, expectedErr, err)
	assert.False(t, sendWasCalled)

	time.Sleep(time.Second)

	select {
	case argClose := <-arg.ChanGracefullyClose:
		assert.Equal(t, common.ImportComplete, argClose.Reason)
	default:
		assert.Fail(t, "did not wrote on end chan")
	}
}

func TestSliceRequester_RequestDataFromHashMarshalFails(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	arg.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedErr
		},
	}
	arg.Storage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			assert.Fail(t, "should not have been called")
			return nil
		},
	}
	sr, _ := NewSliceRequester(arg)

	err := sr.RequestDataFromHash([]byte("hash"), 0)
	assert.Equal(t, expectedErr, err)
}

func TestSliceRequester_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	sendWasCalled := false
	arg := createMockSliceRequesterArg()
	arg.Storage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendWasCalled = true
			return nil
		},
	}
	sr, _ := NewSliceRequester(arg)

	err := sr.RequestDataFromHash([]byte("hash"), 0)

	assert.Nil(t, err)
	assert.True(t, sendWasCalled)
}

func TestSliceRequester_RequestDataFromHashesPackDataInChunksFails(t *testing.T) {
	t.Parallel()

	numGetCalled := 0
	arg := createMockSliceRequesterArg()
	arg.Storage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			numGetCalled++
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			assert.Fail(t, "should not have been called")
			return nil
		},
	}
	arg.DataPacker = &mock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			return nil, expectedErr
		},
	}
	sr, _ := NewSliceRequester(arg)

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	err := sr.RequestDataFromHashArray(hashes, 0)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, len(hashes), numGetCalled)
}

func TestSliceRequester_RequestDataFromHashesShouldWork(t *testing.T) {
	t.Parallel()

	numSendCalled := 0
	numGetCalled := 0
	arg := createMockSliceRequesterArg()
	arg.Storage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			numGetCalled++
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			numSendCalled++
			return nil
		},
	}
	sr, _ := NewSliceRequester(arg)

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	err := sr.RequestDataFromHashArray(hashes, 0)

	assert.Nil(t, err)
	assert.Equal(t, len(hashes), numSendCalled)
	assert.Equal(t, len(hashes), numGetCalled)
}

func TestSliceRequester_GetErroredShouldReturnErr(t *testing.T) {
	t.Parallel()

	numSendCalled := 0
	numGetCalled := 0
	arg := createMockSliceRequesterArg()
	arg.Storage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			numGetCalled++
			if numGetCalled == 1 {
				return nil, expectedErr
			}

			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			numSendCalled++
			return nil
		},
	}
	arg.ChanGracefullyClose = make(chan endProcess.ArgEndProcess, 1)
	sr, _ := NewSliceRequester(arg)

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	err := sr.RequestDataFromHashArray(hashes, 0)

	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, len(hashes)-1, numSendCalled)
	assert.Equal(t, len(hashes), numGetCalled)

	time.Sleep(time.Second)

	select {
	case argClose := <-arg.ChanGracefullyClose:
		assert.Equal(t, common.ImportComplete, argClose.Reason)
	default:
		assert.Fail(t, "did not wrote on end chan")
	}
}

func TestSliceRequester_SendErroredShouldReturnErr(t *testing.T) {
	t.Parallel()

	numSendCalled := 0
	numGetCalled := 0
	arg := createMockSliceRequesterArg()
	arg.Storage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			numGetCalled++
			return make([]byte, 0), nil
		},
	}
	arg.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			numSendCalled++
			if numSendCalled == 1 {
				return expectedErr
			}
			return nil
		},
	}
	sr, _ := NewSliceRequester(arg)

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	err := sr.RequestDataFromHashArray(hashes, 0)

	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, 1, numSendCalled)
	assert.Equal(t, len(hashes), numGetCalled)
}

func TestSliceRequester_Close(t *testing.T) {
	t.Parallel()

	arg := createMockSliceRequesterArg()
	closeCalled := 0
	arg.Storage = &storageStubs.StorerStub{
		CloseCalled: func() error {
			closeCalled++
			return nil
		},
	}
	sr, _ := NewSliceRequester(arg)

	assert.Nil(t, sr.Close())
	assert.Equal(t, 1, closeCalled)
}
