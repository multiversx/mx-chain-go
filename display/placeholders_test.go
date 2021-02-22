package display_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/stretchr/testify/assert"
)

const (
	testMessage   = "message to display"
	testTimestamp = "2000-01-01 00:00:00"
)

func TestHeadline_MessageTooLongItShouldBeReturnedAsItIs(t *testing.T) {
	t.Parallel()

	message := "message way too long to be displayed in a single line with the timestamp and delimiter included."
	message += "in the moment of writing this test, the line was limited to 100 characters"
	delimiter := "."

	res := display.Headline(message, testTimestamp, delimiter)

	assert.Equal(t, message, res)
}

func TestHeadline_DelimiterTooLongShouldBeTrimmed(t *testing.T) {
	t.Parallel()

	firstCharOnDelimiter := "!"
	delimiter := firstCharOnDelimiter + "~="

	res := display.Headline(testMessage, testTimestamp, delimiter)

	assert.False(t, strings.Contains(res, delimiter))
	assert.True(t, strings.Contains(res, firstCharOnDelimiter))
	fmt.Println(res)
}

func TestHeadline_ResultedStringContainsAllData(t *testing.T) {
	t.Parallel()

	delimiter := "."

	res := display.Headline(testMessage, testTimestamp, delimiter)

	assert.True(t, strings.Contains(res, testMessage))
	assert.True(t, strings.Contains(res, testTimestamp))
	assert.True(t, strings.Contains(res, delimiter))
	fmt.Println(res)
}
