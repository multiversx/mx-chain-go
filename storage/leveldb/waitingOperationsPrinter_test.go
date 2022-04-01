package leveldb

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
	loggerMock "github.com/ElrondNetwork/elrond-go/testscommon/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitingOperationsPrinter_startExecuting(t *testing.T) {
	t.Parallel()

	printer := newWaitingOperationsPrinter()
	printer.startExecuting(common.TestPriority)
	assert.Equal(t, 1, printer.counters.counters[common.TestPriority])
}

func TestWaitingOperationsPrinter_endExecuting(t *testing.T) {
	t.Parallel()

	expectedDebugMessage := "storage problems ended, all pending operations finished"
	operation := "operation"
	t.Run("num non zero counters is 0, low priority and no problems should not print", func(t *testing.T) {
		printer := newWaitingOperationsPrinter()
		printer.startExecuting(common.LowPriority)
		printer.logger = &loggerMock.LoggerStub{
			DebugCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called debug")
			},
			WarnCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called warn")
			},
		}

		printer.endExecuting(common.TestPriority, operation, time.Now().Add(-time.Hour))
	})
	t.Run("num non zero counters is 0 and low priority with problems should not print", func(t *testing.T) {
		printer := newWaitingOperationsPrinter()
		printer.startExecuting(common.LowPriority)
		printer.startExecuting(common.LowPriority)
		printer.endExecuting(common.LowPriority, operation, time.Now().Add(-time.Hour))

		printer.logger = &loggerMock.LoggerStub{
			DebugCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called debug")
			},
			WarnCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called warn")
			},
		}

		printer.endExecuting(common.LowPriority, "operation", time.Now().Add(-time.Hour))
	})
	t.Run("num non zero counters is 0 and high priority with problems should print the debug line", func(t *testing.T) {
		printer := newWaitingOperationsPrinter()
		printer.startExecuting(common.HighPriority)
		printer.startExecuting(common.HighPriority)
		printer.endExecuting(common.HighPriority, operation, time.Now().Add(-time.Hour))

		debugCalled := false
		printer.logger = &loggerMock.LoggerStub{
			DebugCalled: func(message string, args ...interface{}) {
				debugCalled = true
				assert.Equal(t, expectedDebugMessage, message)
			},
			WarnCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called warn")
			},
		}

		printer.endExecuting(common.HighPriority, operation, time.Now())
		assert.True(t, debugCalled)
	})
	t.Run("num non zero counters is 0 and high priority with problems should print the debug line and warn line", func(t *testing.T) {
		printer := newWaitingOperationsPrinter()
		printer.startExecuting(common.HighPriority)
		printer.startExecuting(common.HighPriority)
		printer.endExecuting(common.HighPriority, operation, time.Now().Add(-time.Hour))

		debugCalled := false
		warnCalled := false
		printer.logger = &loggerMock.LoggerStub{
			DebugCalled: func(message string, args ...interface{}) {
				debugCalled = true
				assert.Equal(t, expectedDebugMessage, message)
			},
			WarnCalled: func(message string, args ...interface{}) {
				warnCalled = true
				require.Equal(t, 6, len(args))
				assert.Equal(t, operation, args[3].(string))
			},
		}

		printer.endExecuting(common.HighPriority, "operation", time.Now().Add(-time.Hour))
		assert.True(t, debugCalled)
		assert.True(t, warnCalled)
	})
}
