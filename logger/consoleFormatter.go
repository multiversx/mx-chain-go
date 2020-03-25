package logger

import "fmt"

const (
	ansiRegularGray      = "0;37m"
	ansiRegularLightBlue = "0;36m"
	ansiRegularGreen     = "0;32m"
	ansiRegularYellow    = "0;33m"
	ansiRegularRed       = "0;31m"
	ansiRegularBlack     = "0;30m"
)

// ConsoleFormatter implements formatter interface and is used to format log lines to be written on the console
// It uses ANSI-color for colorized console/terminal output.
type ConsoleFormatter struct {
}

// Output converts the provided LogLineHandler into a slice of bytes ready for output
func (cf *ConsoleFormatter) Output(line LogLineHandler) []byte {
	if line == nil {
		return nil
	}

	level := LogLevel(line.GetLogLevel())
	levelColor := getLevelColor(level)
	timestamp := displayTime(line.GetTimestamp())
	loggerName := ""
	correlation := ""
	message := formatMessage(line.GetMessage())
	args := formatArgs(levelColor, line.GetArgs()...)

	if IsEnabledLoggerName() {
		loggerName = formatLoggerName(line.GetLoggerName())
	}

	if IsEnabledCorrelation() {
		correlation = formatCorrelationElements(line.GetCorrelation())
	}

	return []byte(
		fmt.Sprintf(formatColoredString,
			levelColor, level,
			timestamp, loggerName, correlation,
			message, args,
		),
	)
}

// formatArgs iterates through the provided arguments displaying the argument name and after that its value
// The arguments must be provided in the following format: "name1", "val1", "name2", "val2" ...
// It ignores odd number of arguments
func formatArgs(levelColor string, args ...string) string {
	if len(args) == 0 {
		return ""
	}

	argString := ""
	for index := 1; index < len(args); index += 2 {
		argString += fmt.Sprintf("\033[%s%s\033[0m = %s ", levelColor, args[index-1], args[index])
	}

	return argString
}

// getLevelColor output the ANSI color code from the provided log level
func getLevelColor(level LogLevel) string {
	switch level {
	case LogTrace:
		return ansiRegularGray
	case LogDebug:
		return ansiRegularLightBlue
	case LogInfo:
		return ansiRegularGreen
	case LogWarning:
		return ansiRegularYellow
	case LogError:
		return ansiRegularRed
	default:
		return ansiRegularBlack
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (cf *ConsoleFormatter) IsInterfaceNil() bool {
	return cf == nil
}
