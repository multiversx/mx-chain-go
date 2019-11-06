package logger

import (
	"strings"
	"time"
)

const msgFixedLength = 40

func displayTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func formatMessage(msg string) string {
	numWhiteSpaces := 0
	if len(msg) < msgFixedLength {
		numWhiteSpaces = msgFixedLength - len(msg)
	}

	return msg + strings.Repeat(" ", numWhiteSpaces)
}
