package interceptors

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

const fromConnectedPeer = "from connected peer"

//------- preProcessMessage

func newBaseDataInterceptorForPreProcess(throttler process.InterceptorThrottler, antifloodHandler process.P2PAntifloodHandler) *baseDataInterceptor {
	return &baseDataInterceptor{
		throttler:        throttler,
		antifloodHandler: antifloodHandler,
	}
}

func newBaseDataInterceptorForProcess(processor process.InterceptorProcessor, debugHandler process.InterceptedDebugger, topic string) *baseDataInterceptor {
	return &baseDataInterceptor{
		topic:        topic,
		processor:    processor,
		debugHandler: debugHandler,
	}
}

func TestPreProcessMessage_NilMessageShouldErr(t *testing.T) {
	t.Parallel()

	bdi := newBaseDataInterceptorForPreProcess(&mock.InterceptorThrottlerStub{}, &mock.P2PAntifloodHandlerStub{})
	err := bdi.preProcessMesage(nil, fromConnectedPeer)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestPreProcessMessage_NilDataShouldErr(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{}
	bdi := newBaseDataInterceptorForPreProcess(&mock.InterceptorThrottlerStub{}, &mock.P2PAntifloodHandlerStub{})
	err := bdi.preProcessMesage(msg, fromConnectedPeer)

	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestPreProcessMessage_AntifloodCanNotProcessShouldErr(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to process"),
	}
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return false
		},
	}
	expectedErr := errors.New("expected error")
	antifloodHandler := &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return expectedErr
		},
	}

	bdi := newBaseDataInterceptorForPreProcess(throttler, antifloodHandler)
	err := bdi.preProcessMesage(msg, fromConnectedPeer)

	assert.Equal(t, expectedErr, err)
}

func TestPreProcessMessage_AntifloodTopicCanNotProcessShouldErr(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to process"),
	}
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return false
		},
	}
	expectedErr := errors.New("expected error")
	antifloodHandler := &mock.P2PAntifloodHandlerStub{
		CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
			return expectedErr
		},
	}

	bdi := newBaseDataInterceptorForPreProcess(throttler, antifloodHandler)
	err := bdi.preProcessMesage(msg, fromConnectedPeer)

	assert.Equal(t, expectedErr, err)
}

func TestPreProcessMessage_ThrottlerCanNotProcessShouldErr(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to process"),
	}
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return false
		},
	}
	antifloodHandler := &mock.P2PAntifloodHandlerStub{}

	bdi := newBaseDataInterceptorForPreProcess(throttler, antifloodHandler)
	err := bdi.preProcessMesage(msg, fromConnectedPeer)

	assert.Equal(t, process.ErrSystemBusy, err)
}

func TestPreProcessMessage_CanProcessReturnsNilAndCallsStartProcessing(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to process"),
	}
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	bdi := newBaseDataInterceptorForPreProcess(throttler, &mock.P2PAntifloodHandlerStub{})
	err := bdi.preProcessMesage(msg, fromConnectedPeer)

	assert.Nil(t, err)
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
}

func TestPreProcessMessage_CanProcessFromSelf(t *testing.T) {
	t.Parallel()

	currentPeerID := core.PeerID("current peer ID")

	msg := &mock.P2PMessageMock{
		DataField:      []byte("data to process"),
		FromField:      currentPeerID.Bytes(),
		SignatureField: currentPeerID.Bytes(),
	}
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			assert.Fail(t, "should have not called CanProcessCalled")
			return false
		},
	}
	antifloodHandler := &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			assert.Fail(t, "should have not called CanProcessMessageCalled")
			return nil
		},
		CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
			assert.Fail(t, "should have not called CanProcessMessagesOnTopicCalled")
			return nil
		},
	}
	bdi := newBaseDataInterceptorForPreProcess(throttler, antifloodHandler)
	bdi.currentPeerId = currentPeerID
	err := bdi.preProcessMesage(msg, currentPeerID)

	assert.Nil(t, err)
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
}

//------- processInterceptedData

func TestProcessInterceptedData_NotValidShouldCallDoneAndNotCallProcessed(t *testing.T) {
	t.Parallel()

	processCalled := false
	processor := &mock.InterceptorProcessorStub{
		ValidateCalled: func(data process.InterceptedData) error {
			return errors.New("not valid")
		},
		SaveCalled: func(data process.InterceptedData) error {
			processCalled = true
			return nil
		},
	}

	bdi := newBaseDataInterceptorForProcess(processor, &mock.InterceptedDebugHandlerStub{}, "topic")
	bdi.processInterceptedData(&mock.InterceptedDataStub{}, &mock.P2PMessageMock{})

	assert.False(t, processCalled)
}

func TestProcessInterceptedData_ValidShouldCallDoneAndCallProcessed(t *testing.T) {
	t.Parallel()

	processCalled := false
	processor := &mock.InterceptorProcessorStub{
		ValidateCalled: func(data process.InterceptedData) error {
			return nil
		},
		SaveCalled: func(data process.InterceptedData) error {
			processCalled = true
			return nil
		},
	}

	bdi := newBaseDataInterceptorForProcess(processor, &mock.InterceptedDebugHandlerStub{}, "topic")
	bdi.processInterceptedData(&mock.InterceptedDataStub{}, &mock.P2PMessageMock{})

	assert.True(t, processCalled)
}

func TestProcessInterceptedData_ProcessErrorShouldCallDone(t *testing.T) {
	t.Parallel()

	processCalled := false
	processor := &mock.InterceptorProcessorStub{
		ValidateCalled: func(data process.InterceptedData) error {
			return nil
		},
		SaveCalled: func(data process.InterceptedData) error {
			processCalled = true
			return errors.New("error while processing")
		},
	}

	bdi := newBaseDataInterceptorForProcess(processor, &mock.InterceptedDebugHandlerStub{}, "topic")
	bdi.processInterceptedData(&mock.InterceptedDataStub{}, &mock.P2PMessageMock{})

	assert.True(t, processCalled)
}

//------- debug

func TestProcessDebugInterceptedData_ShouldWork(t *testing.T) {
	t.Parallel()

	numCalled := 0
	dh := &mock.InterceptedDebugHandlerStub{
		LogProcessedHashesCalled: func(topic string, hashes [][]byte, err error) {
			numCalled += len(hashes)
		},
	}

	numCalls := 40
	ids := &mock.InterceptedDataStub{
		IdentifiersCalled: func() [][]byte {
			return make([][]byte, numCalls)
		},
	}

	bdi := &baseDataInterceptor{
		debugHandler: dh,
	}
	bdi.processDebugInterceptedData(ids, nil)
	assert.Equal(t, numCalls, numCalled)
}

func TestReceivedDebugInterceptedData_ShouldWork(t *testing.T) {
	t.Parallel()

	numCalled := 0
	dh := &mock.InterceptedDebugHandlerStub{
		LogReceivedHashesCalled: func(topic string, hashes [][]byte) {
			numCalled += len(hashes)
		},
	}

	numCalls := 40
	ids := &mock.InterceptedDataStub{
		IdentifiersCalled: func() [][]byte {
			return make([][]byte, numCalls)
		},
	}

	bdi := &baseDataInterceptor{
		debugHandler: dh,
	}
	bdi.receivedDebugInterceptedData(ids)
	assert.Equal(t, numCalls, numCalled)
}
