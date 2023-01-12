package common

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeoutHandler(t *testing.T) {
	t.Parallel()

	th, err := NewTimeoutHandler(time.Second)
	require.False(t, check.IfNil(th))
	require.Nil(t, err)
	require.True(t, th.checkpoint.Unix() > 0)

	th, err = NewTimeoutHandler(-time.Second)
	require.True(t, check.IfNil(th))
	require.True(t, errors.Is(err, ErrInvalidTimeout))

	th, err = NewTimeoutHandler(time.Nanosecond * 999999999)
	require.True(t, check.IfNil(th))
	require.True(t, errors.Is(err, ErrInvalidTimeout))
}

func TestTimeoutHandler_ResetWatchdog(t *testing.T) {
	t.Parallel()

	startTime := time.Unix(11223344, 0)
	currentTime := startTime
	handler := func() time.Time {
		currentTime = currentTime.Add(time.Second)
		return currentTime
	}

	th := NewTimeoutHandlerWithHandlerFunc(time.Second, handler)
	assert.Equal(t, startTime.Add(time.Second), th.checkpoint)
	th.ResetWatchdog()
	assert.Equal(t, startTime.Add(time.Second*2), th.checkpoint)
	th.ResetWatchdog()
	assert.Equal(t, startTime.Add(time.Second*3), th.checkpoint)
}

func TestTimeoutHandler_IsTimeout(t *testing.T) {
	t.Parallel()

	startTime := time.Unix(11223344, 0)
	currentTime := startTime
	handler := func() time.Time {
		currentTime = currentTime.Add(time.Second)
		return currentTime
	}

	th := NewTimeoutHandlerWithHandlerFunc(time.Second, handler)
	assert.False(t, th.IsTimeout()) // edge case test when checkpoint + timeout = current time

	th = NewTimeoutHandlerWithHandlerFunc(time.Nanosecond*999999999, handler)
	assert.True(t, th.IsTimeout())

	th = NewTimeoutHandlerWithHandlerFunc(time.Nanosecond*1000000001, handler)
	assert.False(t, th.IsTimeout())
}
