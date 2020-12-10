package storageResolvers

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

func createMockTrieResolverArguments() ArgTrieResolver {
	return ArgTrieResolver{
		Messenger:                &mock.MessengerStub{},
		ResponseTopicName:        "",
		Marshalizer:              &mock.MarshalizerStub{},
		TrieDataGetter:           &mock.TrieStub{},
		TrieStorageManager:       &mock.TrieStorageManagerStub{},
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess, 1),
		DelayBeforeGracefulClose: 0,
	}
}

func TestNewTrieNodeResolver_InvalidArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockTrieResolverArguments()
	args.Messenger = nil
	tnr, err := NewTrieNodeResolver(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)

	args = createMockTrieResolverArguments()
	args.ManualEpochStartNotifier = nil
	tnr, err = NewTrieNodeResolver(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilManualEpochStartNotifier, err)

	args = createMockTrieResolverArguments()
	args.ChanGracefullyClose = nil
	tnr, err = NewTrieNodeResolver(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilGracefullyCloseChannel, err)

	args = createMockTrieResolverArguments()
	args.TrieStorageManager = nil
	tnr, err = NewTrieNodeResolver(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilTrieStorageManager, err)

	args = createMockTrieResolverArguments()
	args.TrieDataGetter = nil
	tnr, err = NewTrieNodeResolver(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)

	args = createMockTrieResolverArguments()
	args.Marshalizer = nil
	tnr, err = NewTrieNodeResolver(args)
	assert.True(t, check.IfNil(tnr))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewTrieNodeResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieResolverArguments()
	tnr, err := NewTrieNodeResolver(args)
	assert.False(t, check.IfNil(tnr))
	assert.Nil(t, err)
}

func TestTrieNodeResolver_RequestDataFromHashGetSubtrieFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockTrieResolverArguments()
	expectedErr := errors.New("expected error")
	args.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodesCalled: func(bytes []byte, u uint64) ([][]byte, uint64, error) {
			return nil, 0, expectedErr
		},
	}
	tnr, _ := NewTrieNodeResolver(args)

	err := tnr.RequestDataFromHash(nil, 0)
	assert.Equal(t, expectedErr, err)

	select {
	case <-args.ChanGracefullyClose:
	case <-time.After(time.Second):
		assert.Fail(t, "timout while waiting to signal on gracefully close channel")
	}
}

func TestTrieNodeResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieResolverArguments()
	buff := []byte("data")
	args.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodesCalled: func(bytes []byte, u uint64) ([][]byte, uint64, error) {
			return [][]byte{buff}, 1, nil
		},
	}
	numSendToConnectedPeerCalled := uint32(0)
	args.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			atomic.AddUint32(&numSendToConnectedPeerCalled, 1)
			return nil
		},
	}
	args.Marshalizer = &mock.MarshalizerMock{}
	tnr, _ := NewTrieNodeResolver(args)

	err := tnr.RequestDataFromHash(nil, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(args.ChanGracefullyClose))
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numSendToConnectedPeerCalled))
}

func TestTrieNodeResolver_RequestDataFromHashArrayShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieResolverArguments()
	buff := []byte("data")
	numGetSerializedNodesCalled := uint32(0)
	args.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodesCalled: func(bytes []byte, u uint64) ([][]byte, uint64, error) {
			atomic.AddUint32(&numGetSerializedNodesCalled, 1)
			return [][]byte{buff}, 1, nil
		},
	}
	numSendToConnectedPeerCalled := uint32(0)
	args.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			atomic.AddUint32(&numSendToConnectedPeerCalled, 1)
			return nil
		},
	}
	args.Marshalizer = &mock.MarshalizerMock{}
	tnr, _ := NewTrieNodeResolver(args)

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
