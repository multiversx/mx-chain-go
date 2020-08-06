package forking

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewGenericEpochNotifier(t *testing.T) {
	t.Parallel()

	gep := NewGenericEpochNotifier()

	assert.False(t, check.IfNil(gep))
}

func TestGenericEpochNotifier_RegisterNotifyHandlerNilHandlerShouldNotAdd(t *testing.T) {
	t.Parallel()

	gep := NewGenericEpochNotifier()

	gep.RegisterNotifyHandler(nil)
	assert.Equal(t, 0, len(gep.Handlers()))
}

func TestGenericEpochNotifier_RegisterNotifyHandlerShouldWork(t *testing.T) {
	t.Parallel()

	gep := NewGenericEpochNotifier()

	initialConfirmation := false
	handler := &mock.EpochSubscriberHandlerStub{
		EpochConfirmedCalled: func(epoch uint32) {
			initialConfirmation = true
		},
	}

	gep.RegisterNotifyHandler(handler)
	assert.Equal(t, 1, len(gep.Handlers()))
	assert.True(t, gep.Handlers()[0] == handler) //pointer testing
	assert.True(t, initialConfirmation)
}

func TestGenericEpochNotifier_UnregisterAllShouldWork(t *testing.T) {
	t.Parallel()

	gep := NewGenericEpochNotifier()
	gep.RegisterNotifyHandler(&mock.EpochSubscriberHandlerStub{})
	gep.RegisterNotifyHandler(&mock.EpochSubscriberHandlerStub{})

	assert.Equal(t, 2, len(gep.Handlers()))

	gep.UnRegisterAll()

	assert.Equal(t, 0, len(gep.Handlers()))
}

func TestGenericEpochNotifier_CheckEpochSameEpochShouldNotCall(t *testing.T) {
	t.Parallel()

	gep := NewGenericEpochNotifier()
	numCalls := uint32(0)
	gep.RegisterNotifyHandler(&mock.EpochSubscriberHandlerStub{
		EpochConfirmedCalled: func(epoch uint32) {
			atomic.AddUint32(&numCalls, 1)
		},
	})

	gep.CheckEpoch(0)
	gep.CheckEpoch(0)

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
}

func TestGenericEpochNotifier_CheckEpochShouldCall(t *testing.T) {
	t.Parallel()

	gep := NewGenericEpochNotifier()
	newEpoch := uint32(839843)
	numCalled := uint32(0)
	gep.RegisterNotifyHandler(&mock.EpochSubscriberHandlerStub{
		EpochConfirmedCalled: func(epoch uint32) {
			if epoch == 0 || epoch == newEpoch {
				atomic.AddUint32(&numCalled, 1)
			}
		},
	})

	gep.CheckEpoch(newEpoch)

	assert.Equal(t, uint32(2), atomic.LoadUint32(&numCalled))
	assert.Equal(t, newEpoch, gep.CurrentEpoch())
}

func TestGenericEpochNotifier_CheckEpochInSyncShouldWork(t *testing.T) {
	t.Parallel()

	gep := NewGenericEpochNotifier()
	newEpoch := uint32(839843)

	handlerWait := time.Second
	numCalls := uint32(0)

	handler := &mock.EpochSubscriberHandlerStub{
		EpochConfirmedCalled: func(epoch uint32) {
			time.Sleep(handlerWait)
			atomic.AddUint32(&numCalls, 1)
		},
	}
	gep.RegisterNotifyHandler(handler)

	start := time.Now()
	gep.CheckEpoch(newEpoch)
	end := time.Now()

	assert.Equal(t, uint32(2), atomic.LoadUint32(&numCalls))
	assert.True(t, end.Sub(start) >= handlerWait)
}
