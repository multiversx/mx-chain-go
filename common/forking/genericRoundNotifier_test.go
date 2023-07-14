package forking

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewGenericRoundNotifier(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()

	assert.False(t, check.IfNil(grp))
}

func TestGenericRoundNotifier_RegisterNotifyHandlerNilHandlerShouldNotAdd(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()

	grp.RegisterNotifyHandler(nil)
	assert.Equal(t, 0, len(grp.Handlers()))
}

func TestGenericRoundNotifier_RegisterNotifyHandlerShouldWork(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()

	initialConfirmation := false
	handler := &mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(round uint64, timestamp uint64) {
			initialConfirmation = true
		},
	}

	grp.RegisterNotifyHandler(handler)
	assert.Equal(t, 1, len(grp.Handlers()))
	assert.True(t, grp.Handlers()[0] == handler) //pointer testing
	assert.True(t, initialConfirmation)
}

func TestGenericRoundNotifier_UnregisterAllShouldWork(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()
	grp.RegisterNotifyHandler(&mock.RoundSubscriberHandlerStub{})
	grp.RegisterNotifyHandler(&mock.RoundSubscriberHandlerStub{})

	assert.Equal(t, 2, len(grp.Handlers()))

	grp.UnRegisterAll()

	assert.Equal(t, 0, len(grp.Handlers()))
}

func TestGenericRoundNotifier_CheckRoundNilHeaderNotCall(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()
	numCalls := uint32(0)
	grp.RegisterNotifyHandler(&mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(round uint64, timestamp uint64) {
			atomic.AddUint32(&numCalls, 1)
		},
	})

	grp.CheckRound(nil)

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
	assert.Equal(t, uint64(0), grp.CurrentTimestamp())
}

func TestGenericRoundNotifier_CheckRoundSameRoundShouldNotCall(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()
	numCalls := uint32(0)
	grp.RegisterNotifyHandler(&mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(round uint64, timestamp uint64) {
			atomic.AddUint32(&numCalls, 1)
		},
	})

	grp.CheckRound(&testscommon.HeaderHandlerStub{ //this header will trigger the inner state update
		RoundField:     0,
		TimestampField: 11,
	})
	grp.CheckRound(&testscommon.HeaderHandlerStub{
		RoundField:     0,
		TimestampField: 12,
	})

	assert.Equal(t, uint32(2), atomic.LoadUint32(&numCalls))
	assert.Equal(t, uint64(11), grp.CurrentTimestamp())
}

func TestGenericRoundNotifier_CheckRoundShouldCall(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()
	newRound := uint64(839843)
	newTimestamp := uint64(1389274)
	numCalled := uint32(0)
	grp.RegisterNotifyHandler(&mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(round uint64, timestamp uint64) {
			if round == 0 || round == newRound {
				atomic.AddUint32(&numCalled, 1)
			}
		},
	})

	grp.CheckRound(&testscommon.HeaderHandlerStub{
		RoundField:     newRound,
		TimestampField: newTimestamp,
	})

	assert.Equal(t, uint32(2), atomic.LoadUint32(&numCalled))
	assert.Equal(t, newRound, grp.CurrentRound())
	assert.Equal(t, newTimestamp, grp.CurrentTimestamp())
}

func TestGenericRoundNotifier_CheckRoundInSyncShouldWork(t *testing.T) {
	t.Parallel()

	grp := NewGenericRoundNotifier()
	newRound := uint64(839843)

	handlerWait := time.Second
	numCalls := uint32(0)

	handler := &mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(round uint64, timestamp uint64) {
			time.Sleep(handlerWait)
			atomic.AddUint32(&numCalls, 1)
		},
	}
	grp.RegisterNotifyHandler(handler)

	start := time.Now()
	grp.CheckRound(&testscommon.HeaderHandlerStub{
		RoundField: newRound,
	})
	end := time.Now()

	assert.Equal(t, uint32(2), atomic.LoadUint32(&numCalls))
	assert.True(t, end.Sub(start) >= handlerWait)
}
