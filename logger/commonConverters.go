package logger

import (
	"fmt"
	"strings"
	"time"
)

const msgFixedLength = 40

func displayTime(time time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		time.Year(),
		time.Month(),
		time.Day(),
		time.Hour(),
		time.Minute(),
		time.Second(),
	)
}

func formatMessage(msg string) string {
	numWhiteSpaces := 0
	if len(msg) < msgFixedLength {
		numWhiteSpaces = msgFixedLength - len(msg)
	}

	return msg + strings.Repeat(" ", numWhiteSpaces)
}
