package logging

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// LogErrAsWarnExceptAsDebugIfClosingError logs an error
func LogErrAsWarnExceptAsDebugIfClosingError(logInstance logger.Logger, err error, message string, args ...interface{}) {
	logErrAsLevelExceptAsDebugIfClosingError(logInstance, logger.LogWarning, err, message, args...)
}

// LogErrAsErrorExceptAsDebugIfClosingError logs an error
func LogErrAsErrorExceptAsDebugIfClosingError(logInstance logger.Logger, err error, message string, args ...interface{}) {
	logErrAsLevelExceptAsDebugIfClosingError(logInstance, logger.LogError, err, message, args...)
}

func logErrAsLevelExceptAsDebugIfClosingError(logInstance logger.Logger, logLevel logger.LogLevel, err error, message string, args ...interface{}) {
	if check.IfNil(logInstance) {
		return
	}
	if err == nil {
		return
	}

	if core.IsClosingError(err) {
		logLevel = logger.LogDebug
	}

	logInstance.Log(logLevel, message, args...)
}
