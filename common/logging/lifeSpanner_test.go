package logging

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLifeSpanner_CloseShouldStopNotifying(t *testing.T) {
	t.Parallel()

	chNotify := make(chan struct{})
	spanner := newLifeSpanner(chNotify, trueCheckHandler, time.Second)
	spanner.close()

	select {
	case <-chNotify:
		assert.Fail(t, "should have not notify")
	case <-time.After(time.Second * 5):
	}
}

func TestLifeSpanner_ShouldProcess(t *testing.T) {
	t.Parallel()

	t.Run("true handler should notify", func(t *testing.T) {
		t.Parallel()

		chNotify := make(chan struct{})
		numNotified := uint32(0)
		spanner := newLifeSpanner(chNotify, trueCheckHandler, time.Second)

		ctxTest, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func(ctx context.Context) {
			for {
				select {
				case <-chNotify:
					atomic.AddUint32(&numNotified, 1)
				case <-ctx.Done():
					return
				}
			}
		}(ctxTest)

		time.Sleep(time.Second*3 + time.Millisecond*500) // 3.5s -> counter should be 3
		spanner.close()

		assert.Equal(t, uint32(3), atomic.LoadUint32(&numNotified))
	})
	t.Run("false handler should not notify", func(t *testing.T) {
		t.Parallel()

		chNotify := make(chan struct{})
		numNotified := uint32(0)
		numCheckCalled := uint32(0)
		checkHandler := func() bool {
			atomic.AddUint32(&numCheckCalled, 1)

			return false
		}
		spanner := newLifeSpanner(chNotify, checkHandler, time.Second)

		ctxTest, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func(ctx context.Context) {
			for {
				select {
				case <-chNotify:
					atomic.AddUint32(&numNotified, 1)
				case <-ctx.Done():
					return
				}
			}
		}(ctxTest)

		time.Sleep(time.Second*3 + time.Millisecond*500)
		spanner.close()

		assert.Equal(t, uint32(0), atomic.LoadUint32(&numNotified))
		assert.Equal(t, uint32(3), atomic.LoadUint32(&numCheckCalled))
	})
}

func TestLifeSpanner_ResetShouldWork(t *testing.T) {
	t.Parallel()

	chNotify := make(chan struct{})
	numNotified := uint32(0)
	spanner := newLifeSpanner(chNotify, trueCheckHandler, time.Second)

	ctxTest, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		for {
			select {
			case <-chNotify:
				atomic.AddUint32(&numNotified, 1)
			case <-ctx.Done():
				return
			}
		}
	}(ctxTest)

	go func(ctx context.Context) {
		// this function will call reset twice as fast as the reset duration
		for {
			select {
			case <-time.After(time.Millisecond * 500):
				spanner.reset()
			case <-ctx.Done():
				return
			}
		}
	}(ctxTest)

	time.Sleep(time.Second*3 + time.Millisecond*700)
	spanner.close()

	assert.Equal(t, uint32(0), atomic.LoadUint32(&numNotified)) // no calls due to resets before the timer expiration
}

func TestLifeSpanner_ResetDurationShouldWork(t *testing.T) {
	t.Parallel()

	chNotify := make(chan struct{})
	numNotified := uint32(0)
	spanner := newLifeSpanner(chNotify, trueCheckHandler, time.Minute)

	spanner.resetDuration(time.Second)

	ctxTest, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		for {
			select {
			case <-chNotify:
				atomic.AddUint32(&numNotified, 1)
			case <-ctx.Done():
				return
			}
		}
	}(ctxTest)

	time.Sleep(time.Second*3 + time.Millisecond*500) // 3.5s -> counter should be 3
	spanner.close()

	assert.Equal(t, uint32(3), atomic.LoadUint32(&numNotified))
}
