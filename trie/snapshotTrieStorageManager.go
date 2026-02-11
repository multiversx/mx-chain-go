package trie

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

type snapshotTrieStorageManager struct {
	tsm                *trieStorageManager
	mainSnapshotStorer snapshotPruningStorer
	epoch              uint32
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

// GetFromOldEpochsWithoutAddingToCache tries to get the value for the given key from old epochs without adding it to cache
func (stsm *snapshotTrieStorageManager) GetFromOldEpochsWithoutAddingToCache(key []byte, maxEpochToSearchFrom uint32) ([]byte, uint32, error) {
	// test point get during snapshot

	val, epoch, err := stsm.mainSnapshotStorer.GetFromOldEpochsWithoutAddingToCache(key, maxEpochToSearchFrom)
	if core.IsClosingError(err) {
		return nil, 0, err
	}
	if len(val) == 0 {
		return nil, 0, ErrKeyNotFound
	}

	stsm.putInPreviousStorerIfAbsent(key, val, epoch)

	foundInEpoch := maxEpochToSearchFrom
	if epoch.HasValue {
		foundInEpoch = epoch.Value
	}

	return val, foundInEpoch, nil
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
	err := stsm.mainSnapshotStorer.PutInEpoch(key, val, stsm.epoch-1)
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
	return stsm.mainSnapshotStorer.PutInEpochWithoutCache(key, data, stsm.epoch)
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
