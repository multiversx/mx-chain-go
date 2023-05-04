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
