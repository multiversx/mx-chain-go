package triesHolder

import (
	"fmt"
	"math"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-storage-go/lrucache/capacity"
)

var _ common.TriesHolder = (*readOnlyDataTriesHolder)(nil)

// readOnlyDataTriesHolder caches persisted tries without retaining evicted entries
// for a commit. The underlying LRU serializes access to its entries.
type readOnlyDataTriesHolder struct {
	cacher storage.SizedLRUCacheHandler
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

	return &readOnlyDataTriesHolder{cacher: cacher}, nil
}

// Put caches a persisted trie. Evicted tries need no pending-commit retention.
func (dth *readOnlyDataTriesHolder) Put(key []byte, tr common.Trie) {
	if check.IfNil(tr) || len(key) == 0 {
		return
	}

	dth.cacher.AddSized(string(key), tr, int64(tr.SizeInMemory()))
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
	dth.cacher.Remove(string(key))
}

// Reset releases all cached tries when the API view is recreated.
func (dth *readOnlyDataTriesHolder) Reset() {
	dth.cacher.Purge()
}

// IsInterfaceNil returns true if the underlying object is nil.
func (dth *readOnlyDataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
