package logger

import (
	"fmt"
	"strings"
	"time"
)

const msgFixedLengthName = 20
const msgFixedLength = 40
const ellipsisString = "..."

func displayTime(timestamp int64) string {
	t := time.Unix(0, timestamp)

	return t.Format("2006-01-02 15:04:05.000")
}

func formatMessage(msg string) string {
	numWhiteSpaces := 0
	if len(msg) < msgFixedLength {
		numWhiteSpaces = msgFixedLength - len(msg)
	}

	return msg + strings.Repeat(" ", numWhiteSpaces)
}

func formatLoggerName(name string) string {
	numWhiteSpaces := 0
	if len(name) < msgFixedLengthName {
		numWhiteSpaces = msgFixedLengthName - len(name)
	}
	if len(name) > msgFixedLengthName {
		startingIndex := len(name) - msgFixedLengthName + len(ellipsisString)
		name = ellipsisString + name[startingIndex:]
	}

	return fmt.Sprintf("[%s]", name) + strings.Repeat(" ", numWhiteSpaces)
}
