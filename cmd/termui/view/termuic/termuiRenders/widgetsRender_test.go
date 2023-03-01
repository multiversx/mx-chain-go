package termuiRenders

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWidgetsRender_prepareLogLines(t *testing.T) {
	t.Parallel()

	wr := &WidgetsRender{}
	numLogLines := 25
	testLogLines25 := make([]string, 0, numLogLines)
	for i := 0; i < numLogLines; i++ {
		testLogLines25 = append(testLogLines25, fmt.Sprintf("test line %d", i))
	}

	t.Run("small size should return empty", func(t *testing.T) {
		t.Parallel()

		for i := 0; i <= 2; i++ {
			result := wr.prepareLogLines(testLogLines25, i)
			assert.Empty(t, result)
		}
	})
	t.Run("equal size should return the same slice", func(t *testing.T) {
		t.Parallel()

		result := wr.prepareLogLines(testLogLines25, numLogLines+2)
		assert.Equal(t, testLogLines25, result)
	})
	t.Run("should trim", func(t *testing.T) {
		t.Parallel()

		result := wr.prepareLogLines(testLogLines25, numLogLines)
		assert.Equal(t, testLogLines25[2:], result)
		assert.Equal(t, 23, len(result))

		result = wr.prepareLogLines(testLogLines25, 10)
		assert.Equal(t, testLogLines25[17:], result)
		assert.Equal(t, 8, len(result))
	})
}
