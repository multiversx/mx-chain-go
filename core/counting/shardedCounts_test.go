package counting

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedCounts(t *testing.T) {
	counts := NewShardedCounts()
	counts.PutCounts("foo", 42)
	counts.PutCounts("bar", 43)

	total := counts.GetTotal()
	asString := counts.String()

	require.Equal(t, int64(85), total)
	require.True(t, asString == "Total:85; [foo]=42; [bar]=43; " || asString == "Total:85; [bar]=43; [foo]=42; ")
}

func TestShardedCounts_ConcurrentReadsAndWrites(t *testing.T) {
	counts := NewShardedCounts()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := int64(0); i < 100; i++ {
			counts.PutCounts("foo", 42)
			counts.PutCounts("bar", 43)
		}

		wg.Done()
	}()

	go func() {
		for i := 0; i < 100; i++ {
			total := counts.GetTotal()
			assert.True(t, total == 0 || total == 42 || total == 43 || total == 85)
		}

		wg.Done()
	}()

	wg.Wait()

	total := counts.GetTotal()
	asString := counts.String()

	require.Equal(t, int64(85), total)
	require.True(t, asString == "Total:85; [foo]=42; [bar]=43; " || asString == "Total:85; [bar]=43; [foo]=42; ")
}

func TestShardedCounts_IsInterfaceNil(t *testing.T) {
	counts := NewShardedCounts()
	require.False(t, counts.IsInterfaceNil())

	thisIsNil := (*ShardedCounts)(nil)
	require.True(t, thisIsNil.IsInterfaceNil())
}
