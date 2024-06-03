package sovereign

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("outgoing-operations-pool")

type cacheEntry struct {
	data     *sovereign.BridgeOutGoingData
	expireAt time.Time
}

// This is a cache which stores outgoing txs data at their specified hash.
// Each entry in cache has an expiry time. We should delete entries from this cache once the confirmation from the notifier
// is received that the outgoing operation has been sent to main chain.
// An unconfirmed operation is a tx data operation which has been stored in cache for longer than the time to wait for
// unconfirmed outgoing operations.
// The leader of the next round should check if there are any unconfirmed operations and try to resend them.
type outGoingOperationsPool struct {
	mutex   sync.RWMutex
	timeout time.Duration
	cache   map[string]*cacheEntry
}

// NewOutGoingOperationPool creates a new outgoing operation pool able to store data with an expiry time
func NewOutGoingOperationPool(expiryTime time.Duration) *outGoingOperationsPool {
	log.Debug("NewOutGoingOperationPool", "time to wait for unconfirmed outgoing operations", expiryTime)

	return &outGoingOperationsPool{
		timeout: expiryTime,
		cache:   map[string]*cacheEntry{},
	}
}

// Add adds the outgoing txs data at the specified hash in the internal cache
func (op *outGoingOperationsPool) Add(data *sovereign.BridgeOutGoingData) {
	if data == nil {
		return
	}

	log.Debug("outGoingOperationsPool.Add",
		"hash", data.Hash,
		"aggregated sig", data.AggregatedSignature,
		"leader sig", data.LeaderSignature,
	)

	hashStr := string(data.Hash)

	op.mutex.Lock()
	defer op.mutex.Unlock()

	if _, exists := op.cache[hashStr]; exists {
		return
	}

	op.cache[hashStr] = &cacheEntry{
		data:     data,
		expireAt: time.Now().Add(op.timeout),
	}
}

// Get returns the outgoing txs data at the specified hash
func (op *outGoingOperationsPool) Get(hash []byte) *sovereign.BridgeOutGoingData {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	if cachedEntry, exists := op.cache[string(hash)]; exists {
		return cachedEntry.data
	}

	return nil
}

// Delete removes the outgoing tx data at the specified hash
func (op *outGoingOperationsPool) Delete(hash []byte) {
	log.Debug("outGoingOperationsPool.Delete", "hash", hash)

	op.mutex.Lock()
	defer op.mutex.Unlock()

	delete(op.cache, string(hash))
}

// ConfirmOperation will confirm the bridge op hash by deleting the entry in the internal cache(while keeping the order).
// If there are no more operations under the parent hash(hashOfHashes), the whole cached entry will be deleted
func (op *outGoingOperationsPool) ConfirmOperation(hashOfHashes []byte, hash []byte) error {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	cachedEntry, found := op.cache[string(hashOfHashes)]
	if !found {
		return fmt.Errorf("%w, hashOfHashes: %s, bridgeOpHash: %s",
			errHashOfHashesNotFound, hex.EncodeToString(hashOfHashes), hex.EncodeToString(hash))
	}

	err := confirmOutGoingBridgeOpHash(cachedEntry, hash)
	if err != nil {
		return err
	}

	if len(cachedEntry.data.OutGoingOperations) == 0 {
		delete(op.cache, string(hashOfHashes))
	}

	log.Debug("outGoingOperationsPool.ConfirmOperation", "hashOfHashes", hashOfHashes, "hash", hash)
	return nil
}

func confirmOutGoingBridgeOpHash(cachedEntry *cacheEntry, hash []byte) error {
	cacheData := cachedEntry.data
	for idx, outGoingOp := range cacheData.OutGoingOperations {
		if bytes.Equal(outGoingOp.Hash, hash) {
			cacheData.OutGoingOperations = removeElement(cacheData.OutGoingOperations, idx)
			return nil
		}
	}

	return fmt.Errorf("%w, hash: %s", errHashOfBridgeOpNotFound, hex.EncodeToString(hash))
}

func removeElement(slice []*sovereign.OutGoingOperation, index int) []*sovereign.OutGoingOperation {
	copy(slice[index:], slice[index+1:])
	return slice[:len(slice)-1]
}

// GetUnconfirmedOperations returns a list of unconfirmed operations.
// An unconfirmed operation is a tx data operation which has been stored in cache for longer
// than the time to wait for unconfirmed outgoing operations.
// Returned list is sorted based on expiry time.
func (op *outGoingOperationsPool) GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData {
	expiredEntries := make([]cacheEntry, 0)

	op.mutex.Lock()
	for _, entry := range op.cache {
		if time.Now().After(entry.expireAt) {
			expiredEntries = append(expiredEntries, *entry)
		}
	}
	op.mutex.Unlock()

	sort.Slice(expiredEntries, func(i, j int) bool {
		return expiredEntries[i].expireAt.Before(expiredEntries[j].expireAt)
	})

	ret := make([]*sovereign.BridgeOutGoingData, len(expiredEntries))
	for i, entry := range expiredEntries {
		ret[i] = entry.data
	}

	return ret
}

// ResetTimer will reset the internal expiry timer for the provided outgoing operations hashes
func (op *outGoingOperationsPool) ResetTimer(hashes [][]byte) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	for _, hash := range hashes {
		cachedEntry, exists := op.cache[string(hash)]
		if !exists {
			log.Error("outGoingOperationsPool.ResetTimer error", "hash not found", hash)
			continue
		}
		cachedEntry.expireAt = time.Now().Add(op.timeout)
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outGoingOperationsPool) IsInterfaceNil() bool {
	return op == nil
}
