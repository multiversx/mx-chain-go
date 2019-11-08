package logger

import "fmt"

// PlainFormatter implements formatter interface and is used to format log lines to be written in the same form
// as ConsoleFormatter but it doesn't use the ANSI colors (useful when writing to a file, for example)
type PlainFormatter struct {
}

// Output converts the provided LogLine into a slice of bytes ready for output
func (PlainFormatter) Output(line *LogLine) []byte {
	return []byte(fmt.Sprintf("%s[%s] %s %s\n",
		line.LogLevel.String(),
		displayTime(line.Timestamp),
		formatMessage(line.Message),
		formatArgsNoAnsi(line.Args...),
	),
	)
}

// formatArgsNoAnsi iterates through the provided arguments displaying the argument name and after that its value
// The arguments must be provided in the following format: "name1", "val1", "name2", "val2" ...
// It ignores odd number of arguments and it does not use ANSI colors
func formatArgsNoAnsi(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	argString := ""
	for index := 1; index < len(args); index += 2 {
		argString += fmt.Sprintf("%s=%v ", args[index-1], args[index])
	}

	return argString
}
