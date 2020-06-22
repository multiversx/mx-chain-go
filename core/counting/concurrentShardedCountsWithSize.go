package counting

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

var _ CountsWithSize = (*ConcurrentShardedCountsWithSize)(nil)

// ConcurrentShardedCountsWithSize keeps counts (with sizes) for a sharded data structure
type ConcurrentShardedCountsWithSize struct {
	mutex   sync.RWMutex
	byShard map[string]concurrentShardedCountsWithSizeItem
}

type concurrentShardedCountsWithSizeItem struct {
	counter     int64
	sizeInBytes int64
}

// NewConcurrentShardedCountsWithSize creates a new ShardedCountsWithSize
func NewConcurrentShardedCountsWithSize() *ConcurrentShardedCountsWithSize {
	return &ConcurrentShardedCountsWithSize{
		byShard: make(map[string]concurrentShardedCountsWithSizeItem),
	}
}

// PutCounts registers counts for a shard
func (counts *ConcurrentShardedCountsWithSize) PutCounts(shardName string, counter int64, sizeInBytes int64) {
	counts.mutex.Lock()
	counts.byShard[shardName] = concurrentShardedCountsWithSizeItem{counter: counter, sizeInBytes: sizeInBytes}
	counts.mutex.Unlock()
}

// GetTotal gets total count
func (counts *ConcurrentShardedCountsWithSize) GetTotal() int64 {
	counts.mutex.RLock()
	defer counts.mutex.RUnlock()

	total := int64(0)
	for _, item := range counts.byShard {
		total += item.counter
	}
	return total
}

// GetTotalSize gets total size
func (counts *ConcurrentShardedCountsWithSize) GetTotalSize() int64 {
	counts.mutex.RLock()
	defer counts.mutex.RUnlock()

	total := int64(0)
	for _, item := range counts.byShard {
		total += item.sizeInBytes
	}
	return total
}

func (counts *ConcurrentShardedCountsWithSize) String() string {
	var builder strings.Builder

	total := counts.GetTotal()
	totalSize := core.ConvertBytes(uint64(counts.GetTotalSize()))
	_, _ = fmt.Fprintf(&builder, "Total:%d (%s); ", total, totalSize)

	counts.mutex.RLock()
	defer counts.mutex.RUnlock()

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
func (counts *ConcurrentShardedCountsWithSize) IsInterfaceNil() bool {
	return counts == nil
}
