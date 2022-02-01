package storageResolvers

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockHeaderResolverArg() ArgHeaderResolver {
	return ArgHeaderResolver{
		Messenger:         &mock.MessageHandlerStub{},
		ResponseTopicName: "",
		NonceConverter: &mock.Uint64ByteSliceConverterMock{
			ToByteSliceCalled: func(u uint64) []byte {
				return make([]byte, 0)
			},
			ToUint64Called: func(bytes []byte) (uint64, error) {
				return 0, nil
			},
		},
		HdrStorage:               genericMocks.NewStorerMock("Hdr", 0),
		HeadersNoncesStorage:     genericMocks.NewStorerMock("HeadersNonces", 0),
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess),
	}
}

func TestNewHeaderResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	arg.Messenger = nil
	hdRes, err := NewHeaderResolver(arg)

	assert.True(t, check.IfNil(hdRes))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewHeaderResolver_NilHeaderStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	arg.HdrStorage = nil
	hdRes, err := NewHeaderResolver(arg)

	assert.True(t, check.IfNil(hdRes))
	assert.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
}

func TestNewHeaderResolver_NilHeaderNonceStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	arg.HeadersNoncesStorage = nil
	hdRes, err := NewHeaderResolver(arg)

	assert.True(t, check.IfNil(hdRes))
	assert.Equal(t, dataRetriever.ErrNilHeadersNoncesStorage, err)
}

func TestNewHeaderResolver_NilNonceConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	arg.NonceConverter = nil
	hdRes, err := NewHeaderResolver(arg)

	assert.True(t, check.IfNil(hdRes))
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewHeaderResolver_NilManualEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	arg.ManualEpochStartNotifier = nil
	hdRes, err := NewHeaderResolver(arg)

	assert.True(t, check.IfNil(hdRes))
	assert.Equal(t, dataRetriever.ErrNilManualEpochStartNotifier, err)
}

func TestNewHeaderResolver_NilChanShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	arg.ChanGracefullyClose = nil
	hdRes, err := NewHeaderResolver(arg)

	assert.True(t, check.IfNil(hdRes))
	assert.Equal(t, dataRetriever.ErrNilGracefullyCloseChannel, err)
}

func TestNewHeaderResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	hdRes, err := NewHeaderResolver(arg)

	assert.False(t, check.IfNil(hdRes))
	assert.Nil(t, err)
}

func TestHeaderResolver_SetEpochHandlerNilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	hdRes, _ := NewHeaderResolver(arg)

	err := hdRes.SetEpochHandler(nil)

	assert.Equal(t, dataRetriever.ErrNilEpochHandler, err)
}

func TestHeaderResolver_SetEpochHandlerShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	hdRes, _ := NewHeaderResolver(arg)

	handler := &mock.EpochHandlerStub{}
	err := hdRes.SetEpochHandler(handler)

	assert.Nil(t, err)
	assert.True(t, handler == hdRes.epochHandler) // pointer testing
}

func TestHeaderResolver_RequestDataFromHashNotFoundNotBufferedChannelShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	newEpochCalled := false
	sendCalled := false
	arg := createMockHeaderResolverArg()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	arg.ManualEpochStartNotifier = &mock.ManualEpochStartNotifierStub{
		NewEpochCalled: func(epoch uint32) {
			newEpochCalled = true
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendCalled = true

			return nil
		},
	}
	hdRes, _ := NewHeaderResolver(arg)

	err := hdRes.RequestDataFromHash([]byte("hash"), 0)

	assert.Equal(t, expectedErr, err)
	assert.True(t, newEpochCalled)
	assert.False(t, sendCalled)
}

func TestHeaderResolver_RequestDataFromHashNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	newEpochCalled := false
	sendCalled := false
	arg := createMockHeaderResolverArg()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	arg.ManualEpochStartNotifier = &mock.ManualEpochStartNotifierStub{
		NewEpochCalled: func(epoch uint32) {
			newEpochCalled = true
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendCalled = true

			return nil
		},
	}
	arg.ChanGracefullyClose = make(chan endProcess.ArgEndProcess, 1)
	hdRes, _ := NewHeaderResolver(arg)

	err := hdRes.RequestDataFromHash([]byte("hash"), 0)

	assert.Equal(t, expectedErr, err)
	assert.True(t, newEpochCalled)
	assert.False(t, sendCalled)

	time.Sleep(time.Second)

	select {
	case argClose := <-arg.ChanGracefullyClose:
		assert.Equal(t, common.ImportComplete, argClose.Reason)
	default:
		assert.Fail(t, "did not wrote on end chan")
	}
}

func TestHeaderResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	newEpochCalled := false
	sendCalled := false
	arg := createMockHeaderResolverArg()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return make([]byte, 0), nil
		},
	}
	arg.ManualEpochStartNotifier = &mock.ManualEpochStartNotifierStub{
		NewEpochCalled: func(epoch uint32) {
			newEpochCalled = true
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendCalled = true

			return nil
		},
	}
	hdRes, _ := NewHeaderResolver(arg)

	err := hdRes.RequestDataFromHash([]byte("hash"), 0)

	assert.Nil(t, err)
	assert.True(t, newEpochCalled)
	assert.True(t, sendCalled)
}

func TestHeaderResolver_RequestDataFromNonceNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	newEpochCalled := false
	sendCalled := false
	arg := createMockHeaderResolverArg()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return make([]byte, 0), nil
		},
	}
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	arg.ManualEpochStartNotifier = &mock.ManualEpochStartNotifierStub{
		NewEpochCalled: func(epoch uint32) {
			newEpochCalled = true
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendCalled = true

			return nil
		},
	}
	hdRes, _ := NewHeaderResolver(arg)

	err := hdRes.RequestDataFromNonce(1, 0)

	assert.Equal(t, expectedErr, err)
	assert.False(t, newEpochCalled)
	assert.False(t, sendCalled)
}

func TestHeaderResolver_RequestDataFromNonceShouldWork(t *testing.T) {
	t.Parallel()

	epochsCalled := make(map[uint32]struct{})
	sendCalled := false
	arg := createMockHeaderResolverArg()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return make([]byte, 0), nil
		},
	}
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return make([]byte, 0), nil
		},
	}
	arg.ManualEpochStartNotifier = &mock.ManualEpochStartNotifierStub{
		NewEpochCalled: func(epoch uint32) {
			epochsCalled[epoch] = struct{}{}
		},
	}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendCalled = true

			return nil
		},
	}
	hdRes, _ := NewHeaderResolver(arg)

	err := hdRes.RequestDataFromNonce(1, 0)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(epochsCalled))
	_, found := epochsCalled[1]
	assert.True(t, found)
	_, found = epochsCalled[2]
	assert.True(t, found)
	assert.True(t, sendCalled)
}

func TestHeaderResolver_RequestDataFromEpochShouldWork(t *testing.T) {
	t.Parallel()

	sendCalled := false
	epochIdentifier := []byte(core.EpochStartIdentifier(math.MaxUint32))
	arg := createMockHeaderResolverArg()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			assert.Equal(t, epochIdentifier, key)
			return make([]byte, 0), nil
		},
	}
	arg.ManualEpochStartNotifier = &mock.ManualEpochStartNotifierStub{}
	arg.Messenger = &mock.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendCalled = true

			return nil
		},
	}
	hdRes, _ := NewHeaderResolver(arg)

	err := hdRes.RequestDataFromEpoch(epochIdentifier)

	assert.Nil(t, err)
	assert.True(t, sendCalled)
}

func TestHeaderResolver_Close(t *testing.T) {
	t.Parallel()

	arg := createMockHeaderResolverArg()
	closeCalled := 0
	arg.HdrStorage = &storageStubs.StorerStub{
		CloseCalled: func() error {
			closeCalled++
			return nil
		},
	}
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		CloseCalled: func() error {
			closeCalled++
			return nil
		},
	}
	hdRes, _ := NewHeaderResolver(arg)

	assert.Nil(t, hdRes.Close())
	assert.Equal(t, 2, closeCalled)
}
