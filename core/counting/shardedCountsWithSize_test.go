package counting

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/require"
)

func TestShardedCountsWithSize(t *testing.T) {
	counts := NewShardedCountsWithSize()
	counts.PutCounts("foo", 42, core.MegabyteSize)
	counts.PutCounts("bar", 43, core.MegabyteSize*3)

	total := counts.GetTotal()
	asString := counts.String()

	require.Equal(t, int64(85), total)
	require.Equal(t, "Total:85 (4.00 MB); [bar]=43 (3.00 MB); [foo]=42 (1.00 MB); ", asString)
}

func TestShardedCountsWithSize_IsInterfaceNil(t *testing.T) {
	counts := NewShardedCountsWithSize()
	require.False(t, counts.IsInterfaceNil())

	thisIsNil := (*ShardedCountsWithSize)(nil)
	require.True(t, thisIsNil.IsInterfaceNil())
}
