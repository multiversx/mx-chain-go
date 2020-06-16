package counting

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
)

var _ CountsWithSize = (*ShardedCountsWithSize)(nil)

// ShardedCounts keeps counts (with sizes) for a sharded data structure
// This implementation is NOT concurrently safe
type ShardedCountsWithSize struct {
	byShard map[string]shardedCountsWithSizeItem
}

type shardedCountsWithSizeItem struct {
	counter     int64
	sizeInBytes int64
}

// NewShardedCountsWithSize creates a new ShardedCountsWithSize
func NewShardedCountsWithSize() *ShardedCountsWithSize {
	return &ShardedCountsWithSize{
		byShard: make(map[string]shardedCountsWithSizeItem),
	}
}

// PutCounts registers counts for a shard
func (counts *ShardedCountsWithSize) PutCounts(shardName string, counter int64, sizeInBytes int64) {
	counts.byShard[shardName] = shardedCountsWithSizeItem{counter: counter, sizeInBytes: sizeInBytes}
}

// GetTotal gets total count
func (counts *ShardedCountsWithSize) GetTotal() int64 {
	total := int64(0)
	for _, item := range counts.byShard {
		total += item.counter
	}
	return total
}

// GetTotalSize gets total size
func (counts *ShardedCountsWithSize) GetTotalSize() int64 {
	total := int64(0)
	for _, item := range counts.byShard {
		total += item.sizeInBytes
	}
	return total
}

// String returns a string representation of the counts
func (counts *ShardedCountsWithSize) String() string {
	var builder strings.Builder

	total := counts.GetTotal()
	totalSize := core.ConvertBytes(uint64(counts.GetTotalSize()))
	_, _ = fmt.Fprintf(&builder, "Total:%d (%s); ", total, totalSize)

	// First, we sort the keys alphanumerically
	keys := make([]string, 0, len(counts.byShard))
	for key := range counts.byShard {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		item := counts.byShard[key]
		counter := item.counter
		size := core.ConvertBytes(uint64(item.sizeInBytes))
		_, _ = fmt.Fprintf(&builder, "[%s]=%d (%s); ", key, counter, size)
	}

	return builder.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (counts *ShardedCountsWithSize) IsInterfaceNil() bool {
	return counts == nil
}
