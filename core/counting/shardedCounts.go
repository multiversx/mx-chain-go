package counting

import (
	"fmt"
	"strings"
)

var _ Counts = (*ShardedCounts)(nil)

// ShardedCounts keeps counts for a sharded data structure
type ShardedCounts struct {
	total   int64
	byShard map[string]int64
}

// NewShardedCounts creates a new ShardedCounts
func NewShardedCounts() *ShardedCounts {
	return &ShardedCounts{
		byShard: make(map[string]int64),
	}
}

// PutCounts registers counts for a shard
func (counts *ShardedCounts) PutCounts(shardName string, value int64) {
	counts.byShard[shardName] = value
}

// GetTotal gets total count
func (counts *ShardedCounts) GetTotal() int64 {
	total := int64(0)

	for _, count := range counts.byShard {
		total += count
	}

	return total
}

func (counts *ShardedCounts) String() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Total:%d; ", counts.GetTotal())

	for shardName, count := range counts.byShard {
		fmt.Fprintf(&builder, "[%s]=%d; ", shardName, count)
	}

	return builder.String()
}
