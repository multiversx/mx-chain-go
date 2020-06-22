package counting

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentShardedCountsWithSize(t *testing.T) {
	counts := NewConcurrentShardedCountsWithSize()
	counts.PutCounts("foo", 42, core.MegabyteSize)
	counts.PutCounts("bar", 43, core.MegabyteSize*3)

	total := counts.GetTotal()
	asString := counts.String()

	require.Equal(t, int64(85), total)
	require.Equal(t, "Total:85 (4.00 MB); [bar]=43 (3.00 MB); [foo]=42 (1.00 MB); ", asString)
}

func TestConcurrentShardedCountsWithSize_ConcurrentReadsAndWrites(t *testing.T) {
	counts := NewConcurrentShardedCountsWithSize()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := int64(0); i < 100; i++ {
			counts.PutCounts("foo", 42, 128)
			counts.PutCounts("bar", 43, 128)
		}

		wg.Done()
	}()

	go func() {
		for i := 0; i < 100; i++ {
			total := counts.GetTotal()
			totalSize := counts.GetTotalSize()
			assert.True(t, total == 0 || total == 42 || total == 43 || total == 85)
			assert.True(t, totalSize == 0 || totalSize == 128 || totalSize == 256)
		}

		wg.Done()
	}()

	wg.Wait()

	require.Equal(t, int64(85), counts.GetTotal())
	require.Equal(t, int64(256), counts.GetTotalSize())
	require.Equal(t, "Total:85 (256 B); [bar]=43 (128 B); [foo]=42 (128 B); ", counts.String())
}

func TestConcurrentShardedCountsWithSize_IsInterfaceNil(t *testing.T) {
	counts := NewConcurrentShardedCountsWithSize()
	require.False(t, counts.IsInterfaceNil())

	thisIsNil := (*ConcurrentShardedCountsWithSize)(nil)
	require.True(t, thisIsNil.IsInterfaceNil())
}
