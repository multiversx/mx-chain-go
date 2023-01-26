package sender

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/heartbeat/mock"
	"github.com/stretchr/testify/assert"
)

func TestRoutineHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should work concurrently, calling all handlers, twice", func(t *testing.T) {
		t.Parallel()

		ch1 := make(chan time.Time)
		ch2 := make(chan time.Time)
		ch3 := make(chan struct{})

		numExecuteCalled1 := uint32(0)
		numExecuteCalled2 := uint32(0)
		numExecuteCalled3 := uint32(0)

		handler1 := &mock.SenderHandlerStub{
			ExecutionReadyChannelCalled: func() <-chan time.Time {
				return ch1
			},
			ExecuteCalled: func() {
				atomic.AddUint32(&numExecuteCalled1, 1)
			},
		}
		handler2 := &mock.SenderHandlerStub{
			ExecutionReadyChannelCalled: func() <-chan time.Time {
				return ch2
			},
			ExecuteCalled: func() {
				atomic.AddUint32(&numExecuteCalled2, 1)
			},
		}
		handler3 := &mock.HardforkHandlerStub{
			ShouldTriggerHardforkCalled: func() <-chan struct{} {
				return ch3
			},
			ExecuteCalled: func() {
				atomic.AddUint32(&numExecuteCalled3, 1)
			},
		}

		handler := newRoutineHandler(handler1, handler2, handler3)
		handler.delayAfterHardforkMessageBroadcast = time.Second
		time.Sleep(time.Second) // wait for the go routine start

		assert.Equal(t, uint32(1), atomic.LoadUint32(&numExecuteCalled1)) // initial call
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numExecuteCalled2)) // initial call

		go func() {
			time.Sleep(time.Millisecond * 100)
			ch1 <- time.Now()
		}()
		go func() {
			time.Sleep(time.Millisecond * 100)
			ch2 <- time.Now()
		}()
		go func() {
			time.Sleep(time.Millisecond * 100)
			ch3 <- struct{}{}
		}()

		time.Sleep(time.Second * 3) // wait for the iteration

		assert.Equal(t, uint32(2), atomic.LoadUint32(&numExecuteCalled1))
		assert.Equal(t, uint32(2), atomic.LoadUint32(&numExecuteCalled2))
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numExecuteCalled3))
	})
	t.Run("close should work", func(t *testing.T) {
		t.Parallel()

		ch1 := make(chan time.Time)
		ch2 := make(chan time.Time)

		numExecuteCalled1 := uint32(0)
		numExecuteCalled2 := uint32(0)

		numCloseCalled1 := uint32(0)
		numCloseCalled2 := uint32(0)

		handler1 := &mock.SenderHandlerStub{
			ExecutionReadyChannelCalled: func() <-chan time.Time {
				return ch1
			},
			ExecuteCalled: func() {
				atomic.AddUint32(&numExecuteCalled1, 1)
			},
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled1, 1)
			},
		}
		handler2 := &mock.SenderHandlerStub{
			ExecutionReadyChannelCalled: func() <-chan time.Time {
				return ch2
			},
			ExecuteCalled: func() {
				atomic.AddUint32(&numExecuteCalled2, 1)
			},
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled2, 1)
			},
		}
		handler3 := &mock.HardforkHandlerStub{}

		rh := newRoutineHandler(handler1, handler2, handler3)
		time.Sleep(time.Second) // wait for the go routine start

		assert.Equal(t, uint32(1), atomic.LoadUint32(&numExecuteCalled1)) // initial call
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numExecuteCalled2)) // initial call
		assert.Equal(t, uint32(0), atomic.LoadUint32(&numCloseCalled1))
		assert.Equal(t, uint32(0), atomic.LoadUint32(&numCloseCalled2))

		rh.closeProcessLoop()

		time.Sleep(time.Second) // wait for the go routine to stop

		assert.Equal(t, uint32(1), atomic.LoadUint32(&numExecuteCalled1))
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numExecuteCalled2))
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numCloseCalled1))
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numCloseCalled2))
	})
}

func TestRoutineHandler_Close(t *testing.T) {
	t.Parallel()

	t.Run("Close() call should end the go routine", func(t *testing.T) {
		t.Parallel()

		numCloseCalled := uint32(0)
		handler1 := &mock.SenderHandlerStub{
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled, 1)
			},
		}
		handler2 := &mock.SenderHandlerStub{
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled, 1)
			},
		}
		handler3 := &mock.HardforkHandlerStub{
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled, 1)
			},
		}

		rh := newRoutineHandler(handler1, handler2, handler3)

		rh.closeProcessLoop()
		time.Sleep(time.Second)

		assert.Equal(t, uint32(3), atomic.LoadUint32(&numCloseCalled))
	})
	t.Run("Close() call while hardfork handler is triggered should close immediately", func(t *testing.T) {
		t.Parallel()

		ch := make(chan struct{})
		numCloseCalled := uint32(0)
		handler1 := &mock.SenderHandlerStub{
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled, 1)
			},
		}
		handler2 := &mock.SenderHandlerStub{
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled, 1)
			},
		}
		handler3 := &mock.HardforkHandlerStub{
			CloseCalled: func() {
				atomic.AddUint32(&numCloseCalled, 1)
			},
			ShouldTriggerHardforkCalled: func() <-chan struct{} {
				return ch
			},
		}

		go func() {
			ch <- struct{}{}
		}()

		rh := newRoutineHandler(handler1, handler2, handler3)

		time.Sleep(time.Second)

		rh.closeProcessLoop()
		time.Sleep(time.Second)

		assert.Equal(t, uint32(3), atomic.LoadUint32(&numCloseCalled))
	})

}
