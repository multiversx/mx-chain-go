package logger

// LogLevel defines the priority level of a log line. Trace is the lowest priority level, Error is the highest
type LogLevel byte

// These constants are the string representation of the package logging levels.
const (
	LogTrace   LogLevel = 0
	LogDebug   LogLevel = 1
	LogInfo    LogLevel = 2
	LogWarning LogLevel = 3
	LogError   LogLevel = 4
)

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
	default:
		return ""
	}
}
