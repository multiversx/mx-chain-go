package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetClosedUnbufferedChannel(t *testing.T) {
	t.Parallel()

	ch := GetClosedUnbufferedChannel()

	require.True(t, didTriggerHappen(ch))
	require.True(t, didTriggerHappen(ch))
}

func TestUsingUnbufferedChannelForNotifyingEvents(t *testing.T) {
	// this test isn't related to the GetClosedUnbufferedChannel function, but rather demonstrates how
	// unbuffered channels can be used for notifying events
	channelForTriggeringAction := make(chan struct{})

	require.False(t, didTriggerHappen(channelForTriggeringAction))
	require.False(t, didTriggerHappen(channelForTriggeringAction))
	close(channelForTriggeringAction)
	require.True(t, didTriggerHappen(channelForTriggeringAction))
	require.True(t, didTriggerHappen(channelForTriggeringAction))

}

func didTriggerHappen(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestEmptyUint64Channel_EmptyChannel(t *testing.T) {
	t.Parallel()

	ch := make(chan uint64, 5)
	nrReads := EmptyUint64Channel(ch)

	require.Equal(t, 0, nrReads)
}

func TestEmptyUint64Channel_ChannelWithValues(t *testing.T) {
	t.Parallel()

	ch := make(chan uint64, 5)
	ch <- 1
	ch <- 2
	ch <- 3

	nrReads := EmptyUint64Channel(ch)

	require.Equal(t, 3, nrReads)
	require.Equal(t, 0, len(ch))
}

func TestEmptyUint64Channel_FullChannel(t *testing.T) {
	t.Parallel()

	ch := make(chan uint64, 3)
	ch <- 10
	ch <- 20
	ch <- 30

	nrReads := EmptyUint64Channel(ch)

	require.Equal(t, 3, nrReads)
	require.Equal(t, 0, len(ch))
}
