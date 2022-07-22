package logging

import (
	"errors"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestLogErrAsLevelExceptAsDebugIfClosingError(t *testing.T) {
	t.Run("not a closing error", func(t *testing.T) {
		logCalled := false
		log := &testscommon.LoggerStub{
			LogCalled: func(logLevel logger.LogLevel, message string, args ...interface{}) {
				assert.Equal(t, logger.LogWarning, logLevel)
				assert.Equal(t, "test", message)
				assert.Equal(t, []interface{}{"err", "test error", "a", 7, "b", []byte("hash")}, args)

				logCalled = true
			},
		}

		logErrAsLevelExceptAsDebugIfClosingError(log, logger.LogWarning, errors.New("test error"), "test", "a", 7, "b", []byte("hash"))
		assert.True(t, logCalled)
	})

	t.Run("a closing error", func(t *testing.T) {
		logCalled := false
		log := &testscommon.LoggerStub{
			LogCalled: func(logLevel logger.LogLevel, message string, args ...interface{}) {
				assert.Equal(t, logger.LogDebug, logLevel)
				assert.Equal(t, "test", message)
				assert.Equal(t, []interface{}{"err", "DB is closed", "a", 7, "b", []byte("hash")}, args)

				logCalled = true
			},
		}

		logErrAsLevelExceptAsDebugIfClosingError(log, logger.LogWarning, errors.New("DB is closed"), "test", "a", 7, "b", []byte("hash"))
		assert.True(t, logCalled)
	})

	t.Run("no panic on bad input", func(t *testing.T) {
		log := logger.GetOrCreate("test")

		logErrAsLevelExceptAsDebugIfClosingError(log, logger.LogError, errors.New("test error"), "", "a", nil)
		logErrAsLevelExceptAsDebugIfClosingError(log, logger.LogError, nil, "", "a", nil)
		logErrAsLevelExceptAsDebugIfClosingError(nil, logger.LogError, errors.New("test error"), "")
	})
}
