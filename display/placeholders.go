package display

import (
	"fmt"
	"strings"
)

const maxHeadlineLength = 100

// Headline will build a headline message given a delimiter string
//  timestamp parameter will be printed before the repeating delimiter
func Headline(message string, timestamp string, delimiter string) string {
	if len(delimiter) > 1 {
		delimiter = delimiter[:1]
	}
	if len(message) >= maxHeadlineLength {
		return message
	}
	delimiterLength := (maxHeadlineLength - len(message)) / 2
	delimiterText := strings.Repeat(delimiter, delimiterLength)
	return fmt.Sprintf("%s %s %s %s", timestamp, delimiterText, message, delimiterText)
}
