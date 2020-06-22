package interceptors

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

const fromConnectedPeer = "from connected peer"

//------- preProcessMessage

func TestPreProcessMessage_NilMessageShouldErr(t *testing.T) {
	t.Parallel()

	err := preProcessMesage(&mock.InterceptorThrottlerStub{}, &mock.P2PAntifloodHandlerStub{}, nil, fromConnectedPeer, "")

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestPreProcessMessage_NilDataShouldErr(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{}
	err := preProcessMesage(&mock.InterceptorThrottlerStub{}, &mock.P2PAntifloodHandlerStub{}, msg, fromConnectedPeer, "")

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

	err := preProcessMesage(throttler, antifloodHandler, msg, fromConnectedPeer, "")

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

	err := preProcessMesage(throttler, antifloodHandler, msg, fromConnectedPeer, "")

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

	err := preProcessMesage(throttler, antifloodHandler, msg, fromConnectedPeer, "")

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
	antifloodHandler := &mock.P2PAntifloodHandlerStub{}
	err := preProcessMesage(throttler, antifloodHandler, msg, fromConnectedPeer, "")

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

	wg := &sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	processInterceptedData(
		processor,
		&mock.InterceptedDebugHandlerStub{},
		&mock.InterceptedDataStub{},
		"topic",
		wg,
		&mock.P2PMessageMock{},
	)

	select {
	case <-chDone:
		assert.False(t, processCalled)
	case <-time.After(time.Second):
		assert.Fail(t, "timeout while waiting for wait group object to be finished")
	}
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

	wg := &sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	processInterceptedData(
		processor,
		&mock.InterceptedDebugHandlerStub{},
		&mock.InterceptedDataStub{},
		"topic",
		wg,
		&mock.P2PMessageMock{},
	)

	select {
	case <-chDone:
		assert.True(t, processCalled)
	case <-time.After(time.Second):
		assert.Fail(t, "timeout while waiting for wait group object to be finished")
	}
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

	wg := &sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	processInterceptedData(
		processor,
		&mock.InterceptedDebugHandlerStub{},
		&mock.InterceptedDataStub{},
		"topic",
		wg,
		&mock.P2PMessageMock{},
	)

	select {
	case <-chDone:
		assert.True(t, processCalled)
	case <-time.After(time.Second):
		assert.Fail(t, "timeout while waiting for wait group object to be finished")
	}
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

	processDebugInterceptedData(dh, ids, "", nil)
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

	receivedDebugInterceptedData(dh, ids, "")
	assert.Equal(t, numCalls, numCalled)
}
