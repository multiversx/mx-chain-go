package logger

import (
	"fmt"
	"strings"
	"time"
)

const bracketsLength = len("[]")
const loggerNameFixedLength = 20
const correlationElementsFixedLength = 10
const messageFixedLength = 40
const ellipsisString = "..."

func displayTime(timestamp int64) string {
	t := time.Unix(0, timestamp)
	return t.Format("2006-01-02 15:04:05.000")
}

func formatMessage(msg string) string {
	return padRight(msg, messageFixedLength)
}

func padRight(str string, maxLength int) string {
	paddingLength := maxLength - len(str)

	if paddingLength > 0 {
		return str + strings.Repeat(" ", paddingLength)
	}

	return str
}

func formatLoggerName(name string) string {
	name = truncatePrefix(name, loggerNameFixedLength-bracketsLength)
	formattedName := fmt.Sprintf("[%s]", name)

	return padRight(formattedName, loggerNameFixedLength)
}

func truncatePrefix(str string, maxLength int) string {
	if len(str) > maxLength {
		startingIndex := len(str) - maxLength + len(ellipsisString)
		return ellipsisString + str[startingIndex:]
	}

	return str
}

func formatCorrelationElements() string {
	epoch := globalCorrelation.getEpoch()
	round := globalCorrelation.getRound()
	subRound := globalCorrelation.getSubRound()
	formattedElements := fmt.Sprintf("[%d/%d/%s]", epoch, round, subRound)

	return padRight(formattedElements, correlationElementsFixedLength)
}
