package counting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShardedCounts(t *testing.T) {
	counts := NewShardedCounts()
	counts.PutCounts("foo", 42)
	counts.PutCounts("bar", 43)

	total := counts.GetTotal()
	asString := counts.String()

	require.Equal(t, int64(85), total)
	require.True(t, asString == "Total:85; foo:42; bar:43; " || asString == "Total:85; bar:42; foo:43; ")
}
