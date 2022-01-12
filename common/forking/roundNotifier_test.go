package forking_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/common/mock"
	"github.com/stretchr/testify/require"
)

func TestRoundNotifier_RegisterNotifyHandler(t *testing.T) {
	flag1 := atomic.Flag{}
	subscriber1 := &mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(uint64) {
			_ = flag1.Set()
		},
	}

	flag2 := atomic.Flag{}
	subscriber2 := &mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(uint64) {
			_ = flag2.Set()
		},
	}

	roundNotifier := forking.NewRoundNotifier()

	roundNotifier.RegisterNotifyHandler(subscriber1)
	roundNotifier.RegisterNotifyHandler(subscriber2)
	roundNotifier.RegisterNotifyHandler(nil)

	require.True(t, flag1.IsSet())
	require.True(t, flag2.IsSet())
}

func TestRoundNotifier_CheckRound(t *testing.T) {
	counter1 := atomic.Counter{}
	subscriber1 := &mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(uint64) {
			counter1.Increment()
		},
	}

	counter2 := atomic.Counter{}
	subscriber2 := &mock.RoundSubscriberHandlerStub{
		RoundConfirmedCalled: func(uint64) {
			counter2.Increment()
		},
	}

	roundNotifier := forking.NewRoundNotifier()

	roundNotifier.RegisterNotifyHandler(subscriber1)
	roundNotifier.RegisterNotifyHandler(subscriber2)
	require.Equal(t, int64(1), counter1.Get())
	require.Equal(t, int64(1), counter1.Get())

	// Check new round, expect all subscriber are notified
	roundNotifier.CheckRound(1)
	require.Equal(t, int64(2), counter1.Get())
	require.Equal(t, int64(2), counter1.Get())

	// Check same round as before, expect no subscriber is notified
	roundNotifier.CheckRound(1)
	require.Equal(t, int64(2), counter1.Get())
	require.Equal(t, int64(2), counter1.Get())
}

func TestRoundNotifier_IsInterfaceNil(t *testing.T) {
	roundNotifier := forking.NewRoundNotifier()
	require.False(t, roundNotifier.IsInterfaceNil())
}
