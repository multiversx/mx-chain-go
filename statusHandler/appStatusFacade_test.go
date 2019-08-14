package statusHandler_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/mock"
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

	_, err := statusHandler.NewAppStatusFacadeWithHandlers(statusHandler.NewNilStatusHandler(),
		statusHandler.NewPrometheusStatusHandler())

	assert.Nil(t, err)
}

func TestAppStatusFacade_IncrementShouldPass(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 1)
	var metricKey = core.MetricSynchronizedRound

	// we create a new facade which contains a prometheus handler in order to test
	appStatusHandlerStub := mock.AppStatusHandlerStub{
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
	var metricKey = core.MetricSynchronizedRound

	// we create a new facade which contains a prometheus handler in order to test
	appStatusHandlerStub := mock.AppStatusHandlerStub{
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
	var metricKey = core.MetricSynchronizedRound

	// we create a new facade which contains a prometheus handler in order to test
	appStatusHandlerStub := mock.AppStatusHandlerStub{
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
	var metricKey = core.MetricSynchronizedRound

	// we create a new facade which contains a prometheus handler in order to test
	appStatusHandlerStub := mock.AppStatusHandlerStub{
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
