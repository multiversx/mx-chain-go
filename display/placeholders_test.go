package display_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/stretchr/testify/assert"
)

func TestHeadline_MessageTooLongItShouldBeReturnedAsItIs(t *testing.T) {
	t.Parallel()

	message := "message way too long to be displayed in a single line with the timestamp and delimiter included."
	message += "in the moment of writing this test, the line was limited to 100 characters"
	timestamp := "2000-01-01 00:00:00"
	delimiter := "."

	res := display.Headline(message, timestamp, delimiter)

	assert.Equal(t, message, res)
}

func TestHeadline_DelimiterTooLongShouldBeTrimmed(t *testing.T) {
	t.Parallel()

	message := "message to display"
	timestamp := "2000-01-01 00:00:00"
	firstCharOnDelimiter := "!"
	delimiter := firstCharOnDelimiter + "~="

	res := display.Headline(message, timestamp, delimiter)

	assert.False(t, strings.Contains(res, delimiter))
	assert.True(t, strings.Contains(res, firstCharOnDelimiter))
	fmt.Println(res)
}

func TestHeadline_ResultedStringContainsAllData(t *testing.T) {
	t.Parallel()

	message := "message to display"
	timestamp := "2000-01-01 00:00:00"
	delimiter := "."

	res := display.Headline(message, timestamp, delimiter)

	assert.True(t, strings.Contains(res, message))
	assert.True(t, strings.Contains(res, timestamp))
	assert.True(t, strings.Contains(res, delimiter))
	fmt.Println(res)
}
