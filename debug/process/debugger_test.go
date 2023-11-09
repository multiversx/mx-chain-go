package process

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
)

func createMockProcessDebugConfig() config.ProcessDebugConfig {
	return config.ProcessDebugConfig{
		Enabled:              true,
		GoRoutinesDump:       true,
		DebuggingLogLevel:    "*:INFO",
		PollingTimeInSeconds: minAcceptedValue,
	}
}

func TestNewProcessDebugger(t *testing.T) {
	t.Parallel()

	t.Run("invalid PollingTimeInSeconds", func(t *testing.T) {
		t.Parallel()

		configs := createMockProcessDebugConfig()
		configs.PollingTimeInSeconds = minAcceptedValue - 1

		debuggerInstance, err := NewProcessDebugger(configs)

		assert.True(t, check.IfNil(debuggerInstance))
		assert.True(t, errors.Is(err, errInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "PollingTimeInSeconds"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		configs := createMockProcessDebugConfig()
		debuggerInstance, err := NewProcessDebugger(configs)

		assert.False(t, check.IfNil(debuggerInstance))
		assert.Nil(t, err)

		_ = debuggerInstance.Close()
	})
}

func TestDebugger_ProcessLoopAndClose(t *testing.T) {
	t.Parallel()

	t.Run("node is starting, go routines dump active, should not trigger", func(t *testing.T) {
		t.Parallel()

		configs := createMockProcessDebugConfig()

		numGoRoutinesDumpHandlerCalls := int32(0)
		numLogChangeHandlerCalls := int32(0)

		debuggerInstance, _ := NewProcessDebugger(configs)
		debuggerInstance.goRoutinesDumpHandler = func() {
			atomic.AddInt32(&numGoRoutinesDumpHandlerCalls, 1)
		}
		debuggerInstance.logChangeHandler = func() {
			atomic.AddInt32(&numLogChangeHandlerCalls, 1)
		}

		time.Sleep(time.Second*3 + time.Millisecond*500)

		assert.Zero(t, atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Zero(t, atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))

		time.Sleep(time.Second * 3)

		assert.Zero(t, atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Zero(t, atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))

		err := debuggerInstance.Close()
		assert.Nil(t, err)

		time.Sleep(time.Second * 3)

		assert.Zero(t, atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Zero(t, atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))
	})
	t.Run("node is syncing, go routines dump active, should not trigger", func(t *testing.T) {
		t.Parallel()

		configs := createMockProcessDebugConfig()

		numGoRoutinesDumpHandlerCalls := int32(0)
		numLogChangeHandlerCalls := int32(0)

		debuggerInstance, _ := NewProcessDebugger(configs)
		debuggerInstance.goRoutinesDumpHandler = func() {
			atomic.AddInt32(&numGoRoutinesDumpHandlerCalls, 1)
		}
		debuggerInstance.logChangeHandler = func() {
			atomic.AddInt32(&numLogChangeHandlerCalls, 1)
		}
		debuggerInstance.SetLastCommittedBlockRound(223)

		time.Sleep(time.Second*1 + time.Millisecond*500)

		assert.Zero(t, atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Zero(t, atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))

		err := debuggerInstance.Close()
		assert.Nil(t, err)

		time.Sleep(time.Second * 3)

		assert.Zero(t, atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Zero(t, atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))
	})
	t.Run("node is running, go routines dump active, should not trigger", func(t *testing.T) {
		t.Parallel()

		configs := createMockProcessDebugConfig()

		numGoRoutinesDumpHandlerCalls := int32(0)
		numLogChangeHandlerCalls := int32(0)

		debuggerInstance, _ := NewProcessDebugger(configs)
		debuggerInstance.goRoutinesDumpHandler = func() {
			atomic.AddInt32(&numGoRoutinesDumpHandlerCalls, 1)
		}
		debuggerInstance.logChangeHandler = func() {
			atomic.AddInt32(&numLogChangeHandlerCalls, 1)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			for i := uint64(0); ; i++ {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Millisecond * 100):
					debuggerInstance.SetLastCommittedBlockRound(i)
				}
			}
		}()

		time.Sleep(time.Second*3 + time.Millisecond*500)

		assert.Equal(t, int32(0), atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Equal(t, int32(0), atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))

		err := debuggerInstance.Close()
		assert.Nil(t, err)

		time.Sleep(time.Second * 3)

		assert.Equal(t, int32(0), atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Equal(t, int32(0), atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))
	})
	t.Run("node is stuck, go routines dump active, should trigger", func(t *testing.T) {
		t.Parallel()

		configs := createMockProcessDebugConfig()

		numGoRoutinesDumpHandlerCalls := int32(0)
		numLogChangeHandlerCalls := int32(0)

		debuggerInstance, _ := NewProcessDebugger(configs)
		debuggerInstance.goRoutinesDumpHandler = func() {
			atomic.AddInt32(&numGoRoutinesDumpHandlerCalls, 1)
		}
		debuggerInstance.logChangeHandler = func() {
			atomic.AddInt32(&numLogChangeHandlerCalls, 1)
		}
		debuggerInstance.SetLastCommittedBlockRound(223)

		time.Sleep(time.Second*3 + time.Millisecond*500)

		assert.Equal(t, int32(2), atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Equal(t, int32(2), atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))

		err := debuggerInstance.Close()
		assert.Nil(t, err)

		time.Sleep(time.Second * 3)

		assert.Equal(t, int32(2), atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Equal(t, int32(2), atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))
	})
	t.Run("node is stuck, go routines dump inactive, should trigger", func(t *testing.T) {
		t.Parallel()

		configs := createMockProcessDebugConfig()
		configs.GoRoutinesDump = false

		numGoRoutinesDumpHandlerCalls := int32(0)
		numLogChangeHandlerCalls := int32(0)

		debuggerInstance, _ := NewProcessDebugger(configs)
		debuggerInstance.goRoutinesDumpHandler = func() {
			atomic.AddInt32(&numGoRoutinesDumpHandlerCalls, 1)
		}
		debuggerInstance.logChangeHandler = func() {
			atomic.AddInt32(&numLogChangeHandlerCalls, 1)
		}
		debuggerInstance.SetLastCommittedBlockRound(223)

		time.Sleep(time.Second*3 + time.Millisecond*500)

		assert.Equal(t, int32(2), atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Equal(t, int32(0), atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))

		err := debuggerInstance.Close()
		assert.Nil(t, err)

		time.Sleep(time.Second * 3)

		assert.Equal(t, int32(2), atomic.LoadInt32(&numLogChangeHandlerCalls))
		assert.Equal(t, int32(0), atomic.LoadInt32(&numGoRoutinesDumpHandlerCalls))
	})
}
