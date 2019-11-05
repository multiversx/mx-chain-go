package logger

import "fmt"

// ConsoleFormatter implements formatter interface and is used to format log lines to be written on the console
// It uses ANSI-color for colorized console/terminal output.
type ConsoleFormatter struct {
}

// Output converts the provided LogLine into a slice of bytes ready for output
func (ConsoleFormatter) Output(line *LogLine) []byte {
	levelColor := getLevelColor(line.LogLevel)

	return []byte(fmt.Sprintf("\033[%s%s\033[0m[%s] %s %s\n",
		levelColor,
		line.LogLevel.String(),
		displayTime(line.Timestamp),
		formatMessage(line.Message),
		displayArgs(levelColor, line.Args...),
	),
	)
}

// displayArgs iterates through the provided arguments displaying the argument name and after that its value
// The arguments must be provided in the following format: "name1", "val1", "name2", "val2" ...
// It ignores odd number of arguments
func displayArgs(levelColor string, args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	argString := ""
	for index := 1; index < len(args); index += 2 {
		argString += fmt.Sprintf("\033[%s%s\033[0m=%v ", levelColor, args[index-1], args[index])
	}
	return argString
}

// getLevelColor output the ANSI color code from the provided log level
func getLevelColor(level LogLevel) string {
	switch level {
	case LogTrace:
		return "0;37m"
	case LogDebug:
		return "0;36m"
	case LogInfo:
		return "0;32m"
	case LogWarning:
		return "0;33m"
	case LogError:
		return "0;31m"
	default:
		return "0;30m"
	}
}
