package common

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestErrFromChannel(t *testing.T) {
	t.Parallel()

	t.Run("empty channel, should return nil", func(t *testing.T) {
		t.Parallel()

		t.Run("unbuffered chan", func(t *testing.T) {
			t.Parallel()

			errChan := make(chan error)
			assert.Nil(t, ErrFromChan(errChan))
		})

		t.Run("buffered chan", func(t *testing.T) {
			t.Parallel()

			errChan := make(chan error, 1)
			assert.Nil(t, ErrFromChan(errChan))
		})
	})

	t.Run("non empty channel, should return error", func(t *testing.T) {
		t.Parallel()

		t.Run("unbuffered chan", func(t *testing.T) {
			t.Parallel()

			expectedErr := errors.New("expected error")
			errChan := make(chan error)
			go func() {
				errChan <- expectedErr
			}()

			time.Sleep(time.Second) // allow the go routine to start

			assert.Equal(t, expectedErr, ErrFromChan(errChan))
		})

		t.Run("buffered chan", func(t *testing.T) {
			t.Parallel()

			for i := 1; i < 10; i++ {
				errChan := make(chan error, i)
				expectedErr := errors.New("expected error")
				for j := 0; j < i; j++ {
					errChan <- expectedErr
				}

				assert.Equal(t, expectedErr, ErrFromChan(errChan))
			}
		})
	})
}
