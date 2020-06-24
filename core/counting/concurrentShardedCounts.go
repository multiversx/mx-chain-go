package counting

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

var _ Counts = (*ConcurrentShardedCounts)(nil)

// ShardedCounts keeps counts for a sharded data structure
// This implementation is concurrently safe
type ConcurrentShardedCounts struct {
	mutex   sync.RWMutex
	byShard map[string]int64
}

// NewShardedCounts creates a new ShardedCounts
func NewConcurrentShardedCounts() *ConcurrentShardedCounts {
	return &ConcurrentShardedCounts{
		byShard: make(map[string]int64),
	}
}

// PutCounts registers counts for a shard
func (counts *ConcurrentShardedCounts) PutCounts(shardName string, value int64) {
	counts.mutex.Lock()
	counts.byShard[shardName] = value
	counts.mutex.Unlock()
}

// GetTotal gets total count
func (counts *ConcurrentShardedCounts) GetTotal() int64 {
	counts.mutex.RLock()
	defer counts.mutex.RUnlock()

	total := int64(0)
	for _, count := range counts.byShard {
		total += count
	}

	return total
}

func (counts *ConcurrentShardedCounts) String() string {
	var builder strings.Builder

	_, _ = fmt.Fprintf(&builder, "Total:%d; ", counts.GetTotal())

	counts.mutex.RLock()
	defer counts.mutex.RUnlock()

	// First, we sort the keys alphanumerically
	keys := make([]string, 0, len(counts.byShard))
	for key := range counts.byShard {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		_, _ = fmt.Fprintf(&builder, "[%s]=%d; ", key, counts.byShard[key])
	}

	return builder.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (counts *ConcurrentShardedCounts) IsInterfaceNil() bool {
	return counts == nil
}
