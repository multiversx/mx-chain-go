package triesHolder

import (
	"fmt"
	"math"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-storage-go/lrucache/capacity"
)

var _ common.TriesHolder = (*readOnlyDataTriesHolder)(nil)

// readOnlyDataTriesHolder caches persisted tries without retaining evicted entries
// for a commit. The mutex serializes weight changes with replacement and removal.
type readOnlyDataTriesHolder struct {
	mut          sync.Mutex
	cacher       storage.SizedLRUCacheHandler
	maxTriesSize int64
}

type growthReportingTrie interface {
	state.DataTrie
	GetWithVersionAndSizeOnGrowth(key []byte) (value []byte, depth uint32, version uint32, size int, err error)
}

type trackedReadOnlyTrie struct {
	growthReportingTrie
	holder *readOnlyDataTriesHolder
	key    string
	size   int64 // guarded by holder.mut
}

// NewReadOnlyDataTriesHolder creates a size-limited cache for read-only API views.
func NewReadOnlyDataTriesHolder(maxTriesSize uint64) (*readOnlyDataTriesHolder, error) {
	if maxTriesSize < maxTrieSizeMinValue {
		return nil, fmt.Errorf("%w, provided %d, minimum %d", ErrInvalidMaxTrieSizeValue, maxTriesSize, maxTrieSizeMinValue)
	}

	cacher, err := capacity.NewCapacityLRU(math.MaxInt, int64(maxTriesSize))
	if err != nil {
		return nil, err
	}

	return &readOnlyDataTriesHolder{cacher: cacher, maxTriesSize: int64(maxTriesSize)}, nil
}

// Put caches a persisted trie. Evicted tries need no pending-commit retention.
func (dth *readOnlyDataTriesHolder) Put(key []byte, tr common.Trie) {
	dth.PutAndTrack(key, tr)
}

// PutAndTrack returns the trie to attach to the calling account so that its
// first read and subsequent reads can report memory growth to the cache.
func (dth *readOnlyDataTriesHolder) PutAndTrack(key []byte, tr common.Trie) common.Trie {
	if check.IfNil(tr) || len(key) == 0 {
		return tr
	}
	size := int64(tr.SizeInMemory())
	if size > dth.maxTriesSize {
		// The LRU otherwise keeps a single entry regardless of its size.
		return tr
	}

	cacheKey := string(key)
	if reportingTrie, ok := tr.(growthReportingTrie); ok {
		tr = &trackedReadOnlyTrie{growthReportingTrie: reportingTrie, holder: dth, key: cacheKey, size: size}
	}
	dth.mut.Lock()
	dth.cacher.AddSized(cacheKey, tr, size)
	dth.mut.Unlock()
	return tr
}

func (tr *trackedReadOnlyTrie) Get(key []byte) ([]byte, uint32, error) {
	value, depth, _, err := tr.GetWithVersion(key)
	return value, depth, err
}

func (tr *trackedReadOnlyTrie) GetWithVersion(key []byte) ([]byte, uint32, uint32, error) {
	value, depth, version, size, err := tr.GetWithVersionAndSizeOnGrowth(key)
	if size > 0 {
		tr.holder.updateSize(tr, int64(size))
	}
	return value, depth, version, err
}

func (dth *readOnlyDataTriesHolder) updateSize(tr *trackedReadOnlyTrie, size int64) {
	dth.mut.Lock()
	defer dth.mut.Unlock()
	if size <= tr.size {
		// Concurrent reads can report their sizes out of order.
		return
	}
	current, ok := dth.cacher.Peek(tr.key)
	if !ok || current != tr {
		// A request can outlive eviction, replacement or a view reset.
		return
	}
	if size > dth.maxTriesSize {
		dth.cacher.Remove(tr.key)
		return
	}

	tr.size = size
	dth.cacher.AddSized(tr.key, tr, size)
}

// Get returns a cached trie and updates its LRU position.
func (dth *readOnlyDataTriesHolder) Get(key []byte) common.Trie {
	if len(key) == 0 {
		return nil
	}

	value, ok := dth.cacher.Get(string(key))
	if !ok {
		return nil
	}

	tr, _ := value.(common.Trie)
	return tr
}

// GetAll returns no pending tries because this cache only stores persisted state.
func (dth *readOnlyDataTriesHolder) GetAll() []common.Trie {
	return nil
}

// MarkAllAsCommitted does nothing because API views have no pending commits.
func (dth *readOnlyDataTriesHolder) MarkAllAsCommitted() {}

// MarkAsDirty does nothing because this cache is only used by read-only adapters.
func (dth *readOnlyDataTriesHolder) MarkAsDirty(_ []byte) {}

// Remove deletes a cached trie.
func (dth *readOnlyDataTriesHolder) Remove(key []byte) {
	dth.mut.Lock()
	defer dth.mut.Unlock()
	dth.cacher.Remove(string(key))
}

// Reset releases all cached tries when the API view is recreated.
func (dth *readOnlyDataTriesHolder) Reset() {
	dth.mut.Lock()
	defer dth.mut.Unlock()
	dth.cacher.Purge()
}

// IsInterfaceNil returns true if the underlying object is nil.
func (dth *readOnlyDataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
