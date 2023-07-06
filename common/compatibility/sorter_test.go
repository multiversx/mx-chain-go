package compatibility

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortSlice(t *testing.T) {
	t.Parallel()

	t.Run("0 or negative length should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
			}
		}()

		SortSlice(nil, nil, 0)
		SortSlice(nil, nil, -1)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		stringsUnderTest := []string{"xxx", "aaa", "bbb", "baaaa", "ccc", "abb", "aab", "baa", "caa", "baaa"}
		sortedStrings := []string{"aaa", "aab", "abb", "baa", "baaa", "baaaa", "bbb", "caa", "ccc", "xxx"}

		swap := func(a, b int) {
			stringsUnderTest[a], stringsUnderTest[b] = stringsUnderTest[b], stringsUnderTest[a]
		}
		less := func(a, b int) bool {
			return stringsUnderTest[a] < stringsUnderTest[b]
		}
		SortSlice(swap, less, len(stringsUnderTest))

		assert.Equal(t, sortedStrings, stringsUnderTest)
	})
}
