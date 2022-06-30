package sender

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimerWrapper_createTimerAndExecutionReadyChannel(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		wrapper := &timerWrapper{}
		wrapper.CreateNewTimer(time.Second)
		select {
		case <-wrapper.ExecutionReadyChannel():
			return
		case <-ctx.Done():
			assert.Fail(t, "timeout reached")
		}
	})
	t.Run("double call to should execute, should work", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		wrapper := &timerWrapper{}
		wrapper.CreateNewTimer(time.Second)
		wrapper.CreateNewTimer(time.Second)
		select {
		case <-wrapper.ExecutionReadyChannel():
			return
		case <-ctx.Done():
			assert.Fail(t, "timeout reached")
		}
	})
}

func TestTimerWrapper_Close(t *testing.T) {
	t.Parallel()

	t.Run("close on a nil timer should not panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should have not panicked")
			}
		}()
		wrapper := &timerWrapper{}
		wrapper.Close()
	})
	t.Run("double close on a valid timer should not panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should have not panicked")
			}
		}()
		wrapper := &timerWrapper{}
		wrapper.CreateNewTimer(time.Second)
		wrapper.Close()
		wrapper.Close()
	})
	t.Run("close should stop the timer", func(t *testing.T) {
		wrapper := &timerWrapper{}
		wrapper.CreateNewTimer(time.Second)

		wrapper.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		select {
		case <-wrapper.ExecutionReadyChannel():
			assert.Fail(t, "should have not called execute again")
		case <-ctx.Done():
			return
		}
	})
}

func TestTimerWrapper_ExecutionReadyChannelMultipleTriggers(t *testing.T) {
	t.Parallel()

	wrapper := &timerWrapper{}
	wrapper.CreateNewTimer(time.Second)
	numTriggers := 5
	numExecuted := 0
	for i := 0; i < numTriggers; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		select {
		case <-ctx.Done():
			assert.Fail(t, "timeout reached in iteration")
			cancel()
			return
		case <-wrapper.ExecutionReadyChannel():
			fmt.Printf("iteration %d\n", i)
			numExecuted++
			wrapper.CreateNewTimer(time.Second)
		}

		cancel()
	}

	assert.Equal(t, numTriggers, numExecuted)
}
