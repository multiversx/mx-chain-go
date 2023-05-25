package storagerequesters

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func createMockTrieRequesterArguments() ArgTrieRequester {
	return ArgTrieRequester{
		Messenger:                &p2pmocks.MessengerStub{},
		ResponseTopicName:        "",
		Marshalizer:              &mock.MarshalizerStub{},
		TrieDataGetter:           &trieMock.TrieStub{},
		TrieStorageManager:       &storageManager.StorageManagerStub{},
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess, 1),
		DelayBeforeGracefulClose: 0,
	}
}

func TestNewTrieNodeRequester_InvalidArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockTrieRequesterArguments()
	args.Messenger = nil
	tnr, err := NewTrieNodeRequester(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)

	args = createMockTrieRequesterArguments()
	args.ManualEpochStartNotifier = nil
	tnr, err = NewTrieNodeRequester(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilManualEpochStartNotifier, err)

	args = createMockTrieRequesterArguments()
	args.ChanGracefullyClose = nil
	tnr, err = NewTrieNodeRequester(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilGracefullyCloseChannel, err)

	args = createMockTrieRequesterArguments()
	args.TrieStorageManager = nil
	tnr, err = NewTrieNodeRequester(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilTrieStorageManager, err)

	args = createMockTrieRequesterArguments()
	args.TrieDataGetter = nil
	tnr, err = NewTrieNodeRequester(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)

	args = createMockTrieRequesterArguments()
	args.Marshalizer = nil
	tnr, err = NewTrieNodeRequester(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewTrieNodeRequester_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieRequesterArguments()
	tnr, err := NewTrieNodeRequester(args)
	assert.False(t, check.IfNil(tnr))
	assert.Nil(t, err)
}

func TestTrieNodeRequester_RequestDataFromHashGetSubtrieFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockTrieRequesterArguments()
	expectedErr := errors.New("expected error")
	args.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodesCalled: func(bytes []byte, u uint64) ([][]byte, uint64, error) {
			return nil, 0, expectedErr
		},
	}
	tnr, _ := NewTrieNodeRequester(args)

	err := tnr.RequestDataFromHash(nil, 0)
	assert.Equal(t, expectedErr, err)

	select {
	case <-args.ChanGracefullyClose:
	case <-time.After(time.Second):
		assert.Fail(t, "timout while waiting to signal on gracefully close channel")
	}
}

func TestTrieNodeRequester_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieRequesterArguments()
	buff := []byte("data")
	args.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodesCalled: func(bytes []byte, u uint64) ([][]byte, uint64, error) {
			return [][]byte{buff}, 1, nil
		},
	}
	numSendToConnectedPeerCalled := uint32(0)
	args.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			atomic.AddUint32(&numSendToConnectedPeerCalled, 1)
			return nil
		},
	}
	args.Marshalizer = &mock.MarshalizerMock{}
	tnr, _ := NewTrieNodeRequester(args)

	err := tnr.RequestDataFromHash(nil, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.ChanGracefullyClose))
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numSendToConnectedPeerCalled))
}

func TestTrieNodeRequester_RequestDataFromHashArrayMarshalFails(t *testing.T) {
	t.Parallel()

	args := createMockTrieRequesterArguments()
	buff := []byte("data")
	args.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodesCalled: func(bytes []byte, u uint64) ([][]byte, uint64, error) {
			return [][]byte{buff}, 1, nil
		},
	}
	args.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			assert.Fail(t, "should not have been called")
			return nil
		},
	}
	args.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedErr
		},
	}
	tnr, _ := NewTrieNodeRequester(args)

	err := tnr.RequestDataFromHashArray(
		[][]byte{
			[]byte("hash1"),
			[]byte("hash2"),
		}, 0)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 0, len(args.ChanGracefullyClose))
}

func TestTrieNodeRequester_RequestDataFromHashArrayShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieRequesterArguments()
	buff := []byte("data")
	numGetSerializedNodesCalled := uint32(0)
	args.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodesCalled: func(bytes []byte, u uint64) ([][]byte, uint64, error) {
			atomic.AddUint32(&numGetSerializedNodesCalled, 1)
			return [][]byte{buff}, 1, nil
		},
	}
	numSendToConnectedPeerCalled := uint32(0)
	args.Messenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			atomic.AddUint32(&numSendToConnectedPeerCalled, 1)
			return nil
		},
	}
	args.Marshalizer = &mock.MarshalizerMock{}
	tnr, _ := NewTrieNodeRequester(args)

	err := tnr.RequestDataFromHashArray(
		[][]byte{
			[]byte("hash1"),
			[]byte("hash2"),
		}, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.ChanGracefullyClose))
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numSendToConnectedPeerCalled))
	assert.Equal(t, uint32(2), atomic.LoadUint32(&numGetSerializedNodesCalled))
}

func TestTrieNodeRequester_Close(t *testing.T) {
	t.Parallel()

	t.Run("trieStorageManager.Close error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockTrieRequesterArguments()
		args.TrieStorageManager = &testscommon.StorageManagerStub{
			CloseCalled: func() error {
				return expectedErr
			},
		}
		tnr, _ := NewTrieNodeRequester(args)

		err := tnr.Close()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tnr, _ := NewTrieNodeRequester(createMockTrieRequesterArguments())

		err := tnr.Close()
		assert.NoError(t, err)
	})
}
