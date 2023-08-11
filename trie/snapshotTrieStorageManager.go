package trie

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

type snapshotTrieStorageManager struct {
	*trieStorageManager
	mainSnapshotStorer snapshotPruningStorer
	epoch              uint32
}

func newSnapshotTrieStorageManager(tsm *trieStorageManager, epoch uint32) (*snapshotTrieStorageManager, error) {
	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		return nil, fmt.Errorf("invalid storer, type is %T", tsm.mainStorer)
	}

	return &snapshotTrieStorageManager{
		trieStorageManager: tsm,
		mainSnapshotStorer: storer,
		epoch:              epoch,
	}, nil
}

// Get checks all the storers for the given key, and returns it if it is found
func (stsm *snapshotTrieStorageManager) Get(key []byte) ([]byte, error) {
	stsm.storageOperationMutex.Lock()
	defer stsm.storageOperationMutex.Unlock()

	if stsm.closed {
		log.Debug("snapshotTrieStorageManager get context closing", "key", key)
		return nil, core.ErrContextClosing
	}

	// test point get during snapshot

	val, err := stsm.mainSnapshotStorer.Get(key)
	if core.IsClosingError(err) {
		return nil, err
	}
	if len(val) != 0 {
		return val, nil
	}

	return stsm.getFromOtherStorers(key)
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

// Put adds the given value to the main storer
func (stsm *snapshotTrieStorageManager) Put(key, data []byte) error {
	stsm.storageOperationMutex.Lock()
	defer stsm.storageOperationMutex.Unlock()

	if stsm.closed {
		log.Debug("snapshotTrieStorageManager put context closing", "key", key, "data", data)
		return core.ErrContextClosing
	}

	log.Trace("put hash in snapshot storer", "hash", key, "epoch", stsm.epoch)
	return stsm.mainSnapshotStorer.PutInEpochWithoutCache(key, data, stsm.epoch)
}

// GetFromLastEpoch searches only the last epoch storer for the given key
func (stsm *snapshotTrieStorageManager) GetFromLastEpoch(key []byte) ([]byte, error) {
	stsm.storageOperationMutex.Lock()
	defer stsm.storageOperationMutex.Unlock()

	if stsm.closed {
		log.Debug("snapshotTrieStorageManager getFromLastEpoch context closing", "key", key)
		return nil, core.ErrContextClosing
	}

	return stsm.mainSnapshotStorer.GetFromLastEpoch(key)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stsm *snapshotTrieStorageManager) IsInterfaceNil() bool {
	return stsm == nil
}
