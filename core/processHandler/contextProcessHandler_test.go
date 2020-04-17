package processHandler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var timeout = time.Second

func TestContextProcessHandler_StartGoRoutineLoopingStopabaleGoRoutine(t *testing.T) {
	t.Parallel()

	counter := uint32(0)

	waitTime := time.Millisecond

	cph, _ := NewContextProcessHandler(context.Background())
	cph.StartGoRoutine(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stopped by the cph.Close call")
				return
			default:
				time.Sleep(waitTime)
				atomic.AddUint32(&counter, 1)
			}
		}
	})

	time.Sleep(waitTime * 10)

	_ = cph.Close()

	fmt.Printf("done %d iterations\n", atomic.LoadUint32(&counter))
	assert.True(t, atomic.LoadUint32(&counter) > 0)
}

func TestContextProcessHandler_StartGoRoutineLoopingStopabaleGoRoutineStopShouldNotWait(t *testing.T) {
	t.Parallel()

	counter := uint32(0)
	maxStops := uint32(89)
	counterStop := uint32(0)

	waitTime := time.Millisecond

	cph, _ := NewContextProcessHandler(context.Background())
	cph.StartGoRoutine(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stopping by the cph.Close call")
				time.Sleep(waitTime * 10)
				atomic.StoreUint32(&counterStop, maxStops)

				return
			default:
				time.Sleep(waitTime)
				atomic.AddUint32(&counter, 1)
			}
		}
	})

	time.Sleep(waitTime * 10)

	_ = cph.Close()

	fmt.Printf("done %d iterations\n", atomic.LoadUint32(&counter))
	assert.True(t, atomic.LoadUint32(&counter) > 0)
	assert.NotEqual(t, maxStops, atomic.LoadUint32(&counterStop))
}

func TestContextProcessHandler_StartGoRoutineNestedShouldWork(t *testing.T) {
	t.Parallel()

	counterParent := uint32(0)
	counterChild := uint32(0)
	waitTime := time.Millisecond

	wg := &sync.WaitGroup{}
	wg.Add(2)

	parent, _ := NewContextProcessHandler(context.Background())
	parent.StartGoRoutine(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				wg.Done()

				return
			default:
				time.Sleep(waitTime)
				atomic.AddUint32(&counterParent, 1)
			}
		}
	})

	child, _ := NewContextProcessHandler(parent.Context())
	child.StartGoRoutine(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				wg.Done()

				return
			default:
				time.Sleep(waitTime)
				atomic.AddUint32(&counterChild, 1)
			}
		}
	})

	time.Sleep(waitTime * 10)

	_ = parent.Close()

	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	select {
	case <-time.After(timeout):
		assert.Fail(t, "timeout waiting for process handlers to finish")
	case <-chDone:
	}

	assert.True(t, atomic.LoadUint32(&counterParent) > 0)
	assert.True(t, atomic.LoadUint32(&counterChild) > 0)
}
