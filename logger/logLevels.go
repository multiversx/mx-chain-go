package logger

import (
	"fmt"
	"strings"
)

// LogLevel defines the priority level of a log line. Trace is the lowest priority level, Error is the highest
type LogLevel byte

// These constants are the string representation of the package logging levels.
const (
	LogTrace   LogLevel = 0
	LogDebug   LogLevel = 1
	LogInfo    LogLevel = 2
	LogWarning LogLevel = 3
	LogError   LogLevel = 4
	LogNone    LogLevel = 5
)

// Levels contain all defined levels as a slice for an easier iteration
var Levels = []LogLevel{
	LogTrace,
	LogDebug,
	LogInfo,
	LogWarning,
	LogError,
	LogNone,
}

func (level LogLevel) String() string {
	switch level {
	case LogTrace:
		return "TRACE"
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO "
	case LogWarning:
		return "WARN "
	case LogError:
		return "ERROR"
	case LogNone:
		return "NONE "
	default:
		return ""
	}
}

// GetLogLevel gets the corresponding log level from provided string. The search is case insensitive.
func GetLogLevel(logLevelAsString string) (LogLevel, error) {
	providedLogLevelUpperTrimmed := strings.Trim(strings.ToUpper(logLevelAsString), " ")
	for _, level := range Levels {
		levelUpperTrimmed := strings.Trim(strings.ToUpper(level.String()), " ")
		if levelUpperTrimmed == providedLogLevelUpperTrimmed {
			return level, nil
		}
	}

	return LogTrace, fmt.Errorf("unknown log level provided '%s'", logLevelAsString)
}
