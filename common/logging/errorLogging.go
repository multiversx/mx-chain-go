package logging

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondErrors "github.com/ElrondNetwork/elrond-go/errors"
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

	if elrondErrors.IsClosingError(err) {
		logLevel = logger.LogDebug
	}

	argsWithError := append([]interface{}{"err", err.Error()}, args...)
	logInstance.Log(logLevel, message, argsWithError...)
}
