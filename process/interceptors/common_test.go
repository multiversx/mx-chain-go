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

func TestPreProcessMessage_CanNotProcessShouldErr(t *testing.T) {
	t.Parallel()

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to process"),
	}
	throttler := &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return false
		},
	}

	err := preProcessMesage(throttler, msg)

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
	err := preProcessMesage(throttler, msg)

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

	processInterceptedData(processor, &mock.InterceptedDataStub{}, wg)

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

	processInterceptedData(processor, &mock.InterceptedDataStub{}, wg)

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

	processInterceptedData(processor, &mock.InterceptedDataStub{}, wg)

	select {
	case <-chDone:
		assert.True(t, processCalled)
	case <-time.After(time.Second):
		assert.Fail(t, "timeout while waiting for wait group object to be finished")
	}
}
