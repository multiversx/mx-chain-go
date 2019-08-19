package interceptors

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- preProcessMessage
func TestPreProcessMessage_NilMessageShouldErr(t *testing.T) {
	t.Parallel()

	err := preProcessMesage(&mock.InterceptorThrottlerStub{}, nil)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestPreProcessMessage_NilDataShouldErr(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{}
	err := preProcessMesage(&mock.InterceptorThrottlerStub{}, msg)

	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestPreProcessMessage_CanNotProcessReturnsNil(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to process"),
	}
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessMessageCalled: func() bool {
			return false
		},
	}
	err := preProcessMesage(throttler, msg)

	assert.Nil(t, err)
}

func TestPreProcessMessage_CanProcessReturnsNilAndCallsStartProcessing(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to process"),
	}
	startProcessingCalled := false
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessMessageCalled: func() bool {
			return true
		},
		StartMessageProcessingCalled: func() {
			startProcessingCalled = true
		},
	}
	err := preProcessMesage(throttler, msg)

	assert.Nil(t, err)
	assert.True(t, startProcessingCalled)
}

//------- processInterceptedData

func TestProcessInterceptedData_NotValidShouldCallDoneAndNotCallProcessed(t *testing.T) {
	t.Parallel()

	processCalled := false
	processor := &mock.InterceptorProcessorStub{
		CheckValidForProcessingCalled: func(data process.InterceptedData) error {
			return errors.New("not valid")
		},
		ProcessInteceptedDataCalled: func(data process.InterceptedData) error {
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

	processInterceptedData(processor, mock.InterceptedDataStub{}, wg)

	select {
	case <-chDone:
	case <-time.After(time.Second):
		assert.Fail(t, "timeout while waiting for wait group object to be finished")
	}
	assert.False(t, processCalled)
}

func TestProcessInterceptedData_ValidShouldCallDoneAndCallProcessed(t *testing.T) {
	t.Parallel()

	processCalled := false
	processor := &mock.InterceptorProcessorStub{
		CheckValidForProcessingCalled: func(data process.InterceptedData) error {
			return nil
		},
		ProcessInteceptedDataCalled: func(data process.InterceptedData) error {
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

	processInterceptedData(processor, mock.InterceptedDataStub{}, wg)

	select {
	case <-chDone:
	case <-time.After(time.Second):
		assert.Fail(t, "timeout while waiting for wait group object to be finished")
	}
	assert.True(t, processCalled)
}

func TestProcessInterceptedData_ProcessErrorShouldCallDone(t *testing.T) {
	t.Parallel()

	processCalled := false
	processor := &mock.InterceptorProcessorStub{
		CheckValidForProcessingCalled: func(data process.InterceptedData) error {
			return nil
		},
		ProcessInteceptedDataCalled: func(data process.InterceptedData) error {
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

	processInterceptedData(processor, mock.InterceptedDataStub{}, wg)

	select {
	case <-chDone:
	case <-time.After(time.Second):
		assert.Fail(t, "timeout while waiting for wait group object to be finished")
	}
	assert.True(t, processCalled)
}
