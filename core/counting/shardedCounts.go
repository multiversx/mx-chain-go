package counting

import (
	"fmt"
	"strings"
	"sync"
)

var _ Counts = (*ShardedCounts)(nil)

// ShardedCounts keeps counts for a sharded data structure
// This implementation is concurrently safe
type ShardedCounts struct {
	mutex   sync.RWMutex
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
	counts.mutex.Lock()
	counts.byShard[shardName] = value
	counts.mutex.Unlock()
}

// GetTotal gets total count
func (counts *ShardedCounts) GetTotal() int64 {
	total := int64(0)

	counts.mutex.RLock()

	for _, count := range counts.byShard {
		total += count
	}

	counts.mutex.RUnlock()

	return total
}

// String returns a string representation of the counts
func (counts *ShardedCounts) String() string {
	var builder strings.Builder

	_, _ = fmt.Fprintf(&builder, "Total:%d; ", counts.GetTotal())

	counts.mutex.RLock()

	for shardName, count := range counts.byShard {
		_, _ = fmt.Fprintf(&builder, "[%s]=%d; ", shardName, count)
	}

	counts.mutex.RUnlock()

	return builder.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (counts *ShardedCounts) IsInterfaceNil() bool {
	return counts == nil
}
