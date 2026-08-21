package v2

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	dataRetrieverMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
)

func TestCheckGoRoutinesThrottler(t *testing.T) {
	t.Parallel()

	t.Run("throttler can process should return nil immediately", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		throttler := &dataRetrieverMock.ThrottlerStub{
			CanProcessCalled: func() bool {
				return true
			},
		}

		err := checkGoRoutinesThrottler(ctx, throttler)
		assert.Nil(t, err)
	})

	t.Run("throttler cannot process for a period of time, then can process, should return nil", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		var numCalls int32
		throttler := &dataRetrieverMock.ThrottlerStub{
			CanProcessCalled: func() bool {
				return atomic.AddInt32(&numCalls, 1) >= 5
			},
		}

		err := checkGoRoutinesThrottler(ctx, throttler)
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, atomic.LoadInt32(&numCalls), int32(5))
	})

	t.Run("throttler cannot process should return error when context is done", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		var numCalls int32
		throttler := &dataRetrieverMock.ThrottlerStub{
			CanProcessCalled: func() bool {
				atomic.AddInt32(&numCalls, 1)
				return false
			},
		}

		err := checkGoRoutinesThrottler(ctx, throttler)
		assert.NotNil(t, err)
		assert.ErrorContains(t, err, "time is out")
		assert.ErrorContains(t, err, "while checking the throttler")
		assert.Greater(t, atomic.LoadInt32(&numCalls), int32(1))
	})

	t.Run("context already canceled and throttler cannot process should return error without retrying", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var numCalls int32
		throttler := &dataRetrieverMock.ThrottlerStub{
			CanProcessCalled: func() bool {
				atomic.AddInt32(&numCalls, 1)
				return false
			},
		}

		err := checkGoRoutinesThrottler(ctx, throttler)
		assert.NotNil(t, err)
		assert.ErrorContains(t, err, "time is out")
	})
}
