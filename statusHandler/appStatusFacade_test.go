package statusHandler_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/statusHandler"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewAppStatusFacadeWithHandlers_NilHandlersShouldFail(t *testing.T) {
	t.Parallel()

	_, err := statusHandler.NewAppStatusFacadeWithHandlers()
	assert.Equal(t, statusHandler.ErrHandlersSliceIsNil, err)
}

func TestNewAppStatusFacadeWithHandlers_OneOfTheHandlerIsNilShouldFail(t *testing.T) {
	t.Parallel()

	_, err := statusHandler.NewAppStatusFacadeWithHandlers(statusHandler.NewNilStatusHandler(), nil)
	assert.Equal(t, statusHandler.ErrNilHandlerInSlice, err)
}

func TestNewAppStatusFacadeWithHandlers_OkHandlersShouldPass(t *testing.T) {
	t.Parallel()

	statusFacade, err := statusHandler.NewAppStatusFacadeWithHandlers(
		statusHandler.NewNilStatusHandler(),
		statusHandler.NewNilStatusHandler(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(statusFacade))
}

func TestAppStatusFacade_IncrementShouldPass(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 1)
	var metricKey = common.MetricSynchronizedRound

	// we create a new facade which contains a stub handler in order to test
	appStatusHandlerStub := statusHandlerMock.AppStatusHandlerStub{
		IncrementHandler: func(key string) {
			chanDone <- true
		},
	}

	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(&appStatusHandlerStub)
	assert.Nil(t, err)

	asf.Increment(metricKey)

	select {
	case <-chanDone:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestAppStatusFacade_DecrementShouldPass(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 1)
	var metricKey = common.MetricSynchronizedRound

	// we create a new facade which contains a stub handler in order to test
	appStatusHandlerStub := statusHandlerMock.AppStatusHandlerStub{
		DecrementHandler: func(key string) {
			chanDone <- true
		},
	}

	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(&appStatusHandlerStub)
	assert.Nil(t, err)

	asf.Decrement(metricKey)

	select {
	case <-chanDone:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestAppStatusFacade_SetInt64ValueShouldPass(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 1)
	var metricKey = common.MetricSynchronizedRound

	// we create a new facade which contains a stub handler in order to test
	appStatusHandlerStub := statusHandlerMock.AppStatusHandlerStub{
		SetInt64ValueHandler: func(key string, value int64) {
			chanDone <- true
		},
	}

	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(&appStatusHandlerStub)
	assert.Nil(t, err)

	asf.SetInt64Value(metricKey, int64(0))

	select {
	case <-chanDone:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestAppStatusFacade_SetUint64ValueShouldPass(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 1)
	var metricKey = common.MetricSynchronizedRound

	// we create a new facade which contains a stub handler in order to test
	appStatusHandlerStub := statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			chanDone <- true
		},
	}

	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(&appStatusHandlerStub)
	assert.Nil(t, err)

	asf.SetUInt64Value(metricKey, uint64(0))

	select {
	case <-chanDone:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestAppStatusFacade_AddUint64ShouldPass(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 1)
	var metricKey = common.MetricSynchronizedRound

	// we create a new facade which contains a stub handler in order to test
	appStatusHandlerStub := statusHandlerMock.AppStatusHandlerStub{
		AddUint64Handler: func(key string, value uint64) {
			chanDone <- true
		},
	}

	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(&appStatusHandlerStub)
	assert.Nil(t, err)

	asf.AddUint64(metricKey, 0)

	select {
	case <-chanDone:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestAppStatusFacade_SetStringValueShouldPass(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 1)
	var metricKey = common.MetricNodeDisplayName

	// we create a new facade which contains a stub handler in order to test
	appStatusHandlerStub := statusHandlerMock.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			chanDone <- true
		},
	}

	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(&appStatusHandlerStub)
	assert.Nil(t, err)

	asf.SetStringValue(metricKey, "value")

	select {
	case <-chanDone:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}
