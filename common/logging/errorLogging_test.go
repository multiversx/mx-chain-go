package logging

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func TestLogErrAsLevelExceptAsDebugIfClosingError(t *testing.T) {
	testError := errors.New("test error")
	dbError := errors.New("DB is closed")

	t.Run("not a closing error", func(t *testing.T) {
		logCalled := false
		log := &testscommon.LoggerStub{
			LogCalled: func(logLevel logger.LogLevel, message string, args ...interface{}) {
				assert.Equal(t, logger.LogWarning, logLevel)
				assert.Equal(t, "test", message)
				assert.Equal(t, []interface{}{"a", 7, "b", []byte("hash"), "err", testError.Error()}, args)

				logCalled = true
			},
		}

		LogErrAsWarnExceptAsDebugIfClosingError(log, testError, "test",
			"a", 7,
			"b", []byte("hash"),
			"err", testError.Error(),
		)
		assert.True(t, logCalled)
	})

	t.Run("a closing error", func(t *testing.T) {
		logCalled := false
		log := &testscommon.LoggerStub{
			LogCalled: func(logLevel logger.LogLevel, message string, args ...interface{}) {
				assert.Equal(t, logger.LogDebug, logLevel)
				assert.Equal(t, "test", message)
				assert.Equal(t, []interface{}{"a", 7, "b", []byte("hash"), "err", dbError.Error()}, args)

				logCalled = true
			},
		}

		LogErrAsWarnExceptAsDebugIfClosingError(log, dbError, "test",
			"a", 7,
			"b", []byte("hash"),
			"err", dbError.Error(),
		)
		assert.True(t, logCalled)
	})

	t.Run("no panic on bad input", func(t *testing.T) {
		log := logger.GetOrCreate("test")

		LogErrAsErrorExceptAsDebugIfClosingError(log, testError, "", "a", nil)
		LogErrAsErrorExceptAsDebugIfClosingError(log, nil, "", "a", nil)
		LogErrAsErrorExceptAsDebugIfClosingError(nil, testError, "")
	})
}
