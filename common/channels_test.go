package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetClosedUnbufferedChannel(t *testing.T) {
	t.Parallel()

	ch := GetClosedUnbufferedChannel()

	require.True(t, didTriggerHappened(ch))
	require.True(t, didTriggerHappened(ch))
}

func TestUsingUnbufferedChannelForNotifyingEvents(t *testing.T) {
	// this test isn't related to the GetClosedUnbufferedChannel function, but rather demonstrates how
	// unbuffered channels can be used for notifying events
	channelForTriggeringAction := make(chan struct{})

	require.False(t, didTriggerHappened(channelForTriggeringAction))
	require.False(t, didTriggerHappened(channelForTriggeringAction))
	close(channelForTriggeringAction)
	require.True(t, didTriggerHappened(channelForTriggeringAction))
	require.True(t, didTriggerHappened(channelForTriggeringAction))

}

func didTriggerHappened(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
