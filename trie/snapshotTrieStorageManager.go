package trie

import (
	"bytes"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/snapshotAnalysis"
)

type snapshotTrieStorageManager struct {
	tsm                *trieStorageManager
	mainSnapshotStorer snapshotPruningStorer
	epoch              uint32
	// meter measures the traversal goroutine that owns this instance; nil when the analysis is disabled
	meter *snapshotAnalysis.GoroutineMeter
}

func newSnapshotTrieStorageManager(tsm *trieStorageManager, epoch uint32) (*snapshotTrieStorageManager, error) {
	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		return nil, fmt.Errorf("invalid storer, type is %T", tsm.mainStorer)
	}

	return &snapshotTrieStorageManager{
		tsm:                tsm,
		mainSnapshotStorer: storer,
		epoch:              epoch,
	}, nil
}

// GetWithoutAddingToCache tries to get the value for the given key from old epochs without adding it to cache
func (stsm *snapshotTrieStorageManager) GetWithoutAddingToCache(key []byte, maxEpochToSearchFrom uint32) ([]byte, uint32, error) {
	// test point get during snapshot

	var start time.Time
	if stsm.meter != nil {
		start = time.Now()
	}
	val, epoch, err := stsm.mainSnapshotStorer.GetWithoutAddingToCache(key, maxEpochToSearchFrom)
	if core.IsClosingError(err) {
		return nil, 0, err
	}
	if len(val) == 0 {
		if stsm.meter != nil {
			stsm.meter.GetMiss(time.Since(start))
		}
		return nil, 0, ErrKeyNotFound
	}
	if stsm.meter != nil {
		stsm.meter.Get(time.Since(start), len(val), stsm.foundEpochDistance(epoch))
	}

	stsm.putInPreviousStorerIfAbsent(key, val, epoch)
	foundInEpoch := maxEpochToSearchFrom
	if epoch.HasValue {
		foundInEpoch = epoch.Value
	}

	return val, foundInEpoch, nil
}

// foundEpochDistance returns how many epochs before the snapshot epoch the node was found in, or
// snapshotAnalysis.DistCache for a cache hit (the storer does not know the epoch of a cached value)
func (stsm *snapshotTrieStorageManager) foundEpochDistance(epoch core.OptionalUint32) int {
	if !epoch.HasValue {
		return snapshotAnalysis.DistCache
	}
	if epoch.Value >= stsm.epoch {
		return 0
	}
	return int(stsm.epoch - epoch.Value)
}

func (stsm *snapshotTrieStorageManager) putInPreviousStorerIfAbsent(key []byte, val []byte, epoch core.OptionalUint32) {
	if !epoch.HasValue {
		return
	}

	if stsm.epoch == 0 {
		return
	}

	if epoch.Value >= stsm.epoch-1 {
		return
	}

	if bytes.Equal(key, []byte(common.ActiveDBKey)) || bytes.Equal(key, []byte(common.TrieSyncedKey)) {
		return
	}

	log.Trace("put missing hash in snapshot storer", "hash", key, "epoch", stsm.epoch-1)
	var start time.Time
	if stsm.meter != nil {
		start = time.Now()
	}
	err := stsm.mainSnapshotStorer.PutInEpoch(key, val, stsm.epoch-1)
	if stsm.meter != nil {
		stsm.meter.PrevPut(time.Since(start), len(val))
	}
	if err != nil {
		log.Warn("can not put in epoch",
			"error", err,
			"epoch", stsm.epoch-1,
		)
	}
}

// PutInEpochWithoutCache adds the given (key, data) in the current epoch without adding it to cache
func (stsm *snapshotTrieStorageManager) PutInEpochWithoutCache(key, data []byte) error {
	log.Trace("put hash in snapshot storer", "hash", key, "epoch", stsm.epoch)
	if stsm.meter == nil {
		return stsm.mainSnapshotStorer.PutInEpochWithoutCache(key, data, stsm.epoch)
	}

	start := time.Now()
	err := stsm.mainSnapshotStorer.PutInEpochWithoutCache(key, data, stsm.epoch)
	stsm.meter.Put(time.Since(start), len(data))
	return err
}

// analysisMeter returns the goroutine meter of the traversal that owns this storage manager
func (stsm *snapshotTrieStorageManager) analysisMeter() *snapshotAnalysis.GoroutineMeter {
	return stsm.meter
}

// GetFromLastEpoch searches only the last epoch storer for the given key
func (stsm *snapshotTrieStorageManager) GetFromLastEpoch(key []byte) ([]byte, error) {
	return stsm.mainSnapshotStorer.GetFromLastEpoch(key)
}

// GetIdentifier returns the identifier of the storage manager
func (stsm *snapshotTrieStorageManager) GetIdentifier() string {
	return stsm.tsm.GetIdentifier()
}

// GetFromCurrentEpoch tries to get the value for the given key from the current epoch
func (stsm *snapshotTrieStorageManager) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	return stsm.mainSnapshotStorer.GetFromCurrentEpoch(key)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stsm *snapshotTrieStorageManager) IsInterfaceNil() bool {
	return stsm == nil
}
