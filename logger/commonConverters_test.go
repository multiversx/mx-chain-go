package logger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMessage_ShouldOutputFixedOrLargerStringThanMsgFixedLength(t *testing.T) {
	t.Parallel()

	testData := make(map[string]int)
	testData[""] = 40
	testData["small string"] = 40
	largeString := "large string: " + strings.Repeat("A", msgFixedLength)
	testData[largeString] = len(largeString)

	for k, v := range testData {
		result := formatMessage(k)

		assert.Equal(t, v, len(result))
	}
}
