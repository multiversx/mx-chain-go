package scquery

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	atomicFlags "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsSCQueryDebugger() ArgsSCQueryDebugger {
	return ArgsSCQueryDebugger{
		IntervalAutoPrintInSeconds: 1,
		LoggerInstance:             &testscommon.LoggerStub{},
	}
}

func TestNewSCQueryDebugger(t *testing.T) {
	t.Parallel()

	t.Run("invalid interval should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSCQueryDebugger()
		args.IntervalAutoPrintInSeconds = 0
		debugger, err := NewSCQueryDebugger(args)
		assert.Nil(t, debugger)
		assert.ErrorIs(t, err, errInvalidIntervalAutoPrint)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSCQueryDebugger()
		args.LoggerInstance = nil
		debugger, err := NewSCQueryDebugger(args)
		assert.Nil(t, debugger)
		assert.ErrorIs(t, err, errNilLogger)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSCQueryDebugger()
		debugger, err := NewSCQueryDebugger(args)
		assert.NotNil(t, debugger)
		assert.NoError(t, err)

		_ = debugger.Close()
	})
}

func TestSCQueryDebugger_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var debugger *scQueryDebugger
	require.True(t, debugger.IsInterfaceNil())

	args := createMockArgsSCQueryDebugger()
	debugger, _ = NewSCQueryDebugger(args)
	require.False(t, debugger.IsInterfaceNil())
}

func TestScQueryDebugger_CloseShouldWork(t *testing.T) {
	t.Parallel()

	flagStartCalled := &atomicFlags.Flag{}
	flagStopCalled := &atomicFlags.Flag{}
	args := createMockArgsSCQueryDebugger()
	args.LoggerInstance = &testscommon.LoggerStub{
		TraceCalled: func(message string, args ...interface{}) {
			if message == logStartProcessLoop {
				flagStartCalled.SetValue(true)
			}
		},
		DebugCalled: func(message string, args ...interface{}) {
			if message == logClosedProcessLoop {
				flagStopCalled.SetValue(true)
			}
		},
	}

	debugger, _ := NewSCQueryDebugger(args)
	time.Sleep(time.Millisecond * 100) // wait for the go routine to start
	_ = debugger.Close()
	time.Sleep(time.Millisecond * 100) // wait for the go routine to stop

	assert.True(t, flagStartCalled.IsSet())
	assert.True(t, flagStopCalled.IsSet())
}

func TestScQueryDebugger_ShouldNotPrintIfNoSCQueriesWereStarted(t *testing.T) {
	t.Parallel()

	handler := func(message string, args ...interface{}) {
		if message == logMetrics {
			assert.Fail(t, "should not print metrics")
		}
	}

	args := createMockArgsSCQueryDebugger()
	args.LoggerInstance = &testscommon.LoggerStub{
		TraceCalled: handler,
		DebugCalled: handler,
		InfoCalled:  handler,
		WarnCalled:  handler,
		ErrorCalled: handler,
	}

	debugger, _ := NewSCQueryDebugger(args)
	time.Sleep(time.Second * 2) // wait for the go routine to start & call print

	_ = debugger.Close()
}

func TestScQueryDebugger_ShouldPrintIfSCQueriesWereStarted(t *testing.T) {
	t.Parallel()

	wasCalled := &atomicFlags.Flag{}
	handler := func(message string, args ...interface{}) {
		if message == logMetrics {
			wasCalled.SetValue(true)
		}
	}

	args := createMockArgsSCQueryDebugger()
	args.LoggerInstance = &testscommon.LoggerStub{
		TraceCalled: handler,
		DebugCalled: handler,
		InfoCalled:  handler,
		WarnCalled:  handler,
		ErrorCalled: handler,
	}

	debugger, _ := NewSCQueryDebugger(args)
	debugger.NotifyExecutionStarted(0)

	time.Sleep(time.Second * 2) // wait for the go routine to start & call print

	_ = debugger.Close()

	assert.True(t, wasCalled.IsSet())
}

func TestScQueryDebugger_NotifyShouldWork(t *testing.T) {
	t.Parallel()

	currentTimestamp := int64(5)
	getTimestampHandler := func() int64 {
		return atomic.AddInt64(&currentTimestamp, 5) // each call will get a new time stamp with a difference of 5
	}

	args := createMockArgsSCQueryDebugger()
	args.IntervalAutoPrintInSeconds = 3600 // do not call reset
	debugger, _ := NewTestSCQueryDebugger(args, getTimestampHandler)

	// do not call these tests in parallel, they are using the same debugger instance
	t.Run("should work with a finished call", func(t *testing.T) {
		debugger.FullReset()
		debugger.NotifyExecutionStarted(37)  // timestamp = 10 & 15 (for the recompute metrics)
		debugger.NotifyExecutionFinished(37) // timestamp = 20 (for the recompute metrics)

		expectedResult := Metrics{
			RunningQueries:            make(map[int]int64),
			MaxNumOfConcurrentQueries: 1,
			NumFinishedCalls:          1,
			IndexOfHeavierSCCall:      37,
			DurationOfHeaviestSCCall:  10,
		}

		checkResults(t, expectedResult, debugger.GetMetrics())
	})
	t.Run("should work with a finished call & unfinished call", func(t *testing.T) {
		debugger.FullReset()
		atomic.StoreInt64(&currentTimestamp, 5)

		debugger.NotifyExecutionStarted(37)  // timestamp = 10 & 15 (for the recompute metrics)
		debugger.NotifyExecutionFinished(37) // timestamp = 20 (for the recompute metrics)
		debugger.NotifyExecutionStarted(38)  // timestamp = 25 & 30 (for the recompute metrics)

		expectedResult := Metrics{
			RunningQueries:            map[int]int64{38: 25},
			MaxNumOfConcurrentQueries: 1,
			NumFinishedCalls:          1,
			IndexOfHeavierSCCall:      37,
			DurationOfHeaviestSCCall:  10,
		}

		checkResults(t, expectedResult, debugger.GetMetrics())
	})
	t.Run("should work with an unfinished call", func(t *testing.T) {
		debugger.FullReset()
		atomic.StoreInt64(&currentTimestamp, 5)

		debugger.NotifyExecutionStarted(37) // timestamp = 10 & 15 (for the recompute metrics)
		debugger.Reset()

		expectedResult := Metrics{
			RunningQueries:            map[int]int64{37: 10},
			MaxNumOfConcurrentQueries: 0,
			NumFinishedCalls:          0,
			IndexOfHeavierSCCall:      0,
			DurationOfHeaviestSCCall:  0,
		}

		checkResults(t, expectedResult, debugger.GetMetrics())
	})
	t.Run("should work with 2 unfinished call", func(t *testing.T) {
		debugger.FullReset()
		atomic.StoreInt64(&currentTimestamp, 5)

		debugger.NotifyExecutionStarted(37) // timestamp = 10 & 15 (for the recompute metrics)
		debugger.Reset()

		debugger.NotifyExecutionStarted(38) // timestamp = 20 & 25 (for the recompute metrics)

		expectedResult := Metrics{
			RunningQueries: map[int]int64{
				37: 10,
				38: 20,
			},
			MaxNumOfConcurrentQueries: 2,
			NumFinishedCalls:          0,
			IndexOfHeavierSCCall:      37,
			DurationOfHeaviestSCCall:  15, // current is 25 - 10 for index 37
		}

		checkResults(t, expectedResult, debugger.GetMetrics())
	})
}

func TestScQueryDebugger_LogPrint(t *testing.T) {
	t.Parallel()

	currentTimestamp := int64(5)
	getTimestampHandler := func() int64 {
		return atomic.AddInt64(&currentTimestamp, 5) // each call will get a new time stamp with a difference of 5
	}

	printWasCalled := &atomicFlags.Flag{}
	args := createMockArgsSCQueryDebugger()
	args.IntervalAutoPrintInSeconds = 1
	args.LoggerInstance = &testscommon.LoggerStub{
		DebugCalled: func(message string, argsLogger ...interface{}) {
			if message != logMetrics {
				return
			}

			printWasCalled.SetValue(true)
			expectedArgs := []interface{}{
				logMaxConcurrentQueries, 3,
				logDurationOfHeaviestCall, time.Duration(10), // 10 current - 0 for the 1 index
				logIndexOfHeaviestCall, 1,
				logFinishedCalls, 4,
				logRunningCalls, 2,
				logMetricsInterval, time.Second * time.Duration(args.IntervalAutoPrintInSeconds),
			}

			assert.Equal(t, expectedArgs, argsLogger)
		},
	}
	debugger, _ := NewTestSCQueryDebugger(args, getTimestampHandler)
	debugger.mutData.Lock()
	debugger.maxNumOfConcurrentQueries = 3
	debugger.numFinishedCalls = 4
	debugger.runningQueries = map[int]int64{
		1: 0,
		2: 1,
	}
	debugger.mutData.Unlock()

	time.Sleep(time.Second + time.Millisecond*300) // wait for the print
	assert.True(t, printWasCalled.IsSet())

	_ = debugger.Close()
}

func checkResults(t *testing.T, expected Metrics, actual Metrics) {
	assert.Equal(t, expected.RunningQueries, actual.RunningQueries)
	assert.Equal(t, expected.MaxNumOfConcurrentQueries, actual.MaxNumOfConcurrentQueries)
	assert.Equal(t, expected.NumFinishedCalls, actual.NumFinishedCalls)
	assert.Equal(t, expected.IndexOfHeavierSCCall, actual.IndexOfHeavierSCCall)
	assert.Equal(t, expected.DurationOfHeaviestSCCall, actual.DurationOfHeaviestSCCall)
}

func TestScQueryDebugger_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	numOperations := 1000
	args := createMockArgsSCQueryDebugger()
	debugger, _ := NewSCQueryDebugger(args)
	wg := &sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			time.Sleep(time.Second)

			switch idx % 2 {
			case 0:
				debugger.NotifyExecutionStarted(idx)
			case 1:
				debugger.NotifyExecutionFinished(idx - 1)
			default:
				assert.Fail(t, "wrong setup in test")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	_ = debugger.Close()
}
