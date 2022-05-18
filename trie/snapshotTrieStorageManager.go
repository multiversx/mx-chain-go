package trie

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/errors"
)

type snapshotTrieStorageManager struct {
	pruningStorer
	tsm   *trieStorageManager
	epoch uint32
}

func newSnapshotTrieStorageManager(tsm *trieStorageManager, epoch uint32) (*snapshotTrieStorageManager, error) {
	storer, ok := tsm.mainStorer.(pruningStorer)
	if !ok {
		return nil, fmt.Errorf("invalid storer type")
	}

	return &snapshotTrieStorageManager{
		pruningStorer: storer,
		tsm:           tsm,
		epoch:         epoch,
	}, nil
}

// Put adds the given value to the last epoch storer
func (stsm *snapshotTrieStorageManager) Put(key []byte, val []byte) error {
	if stsm.epoch == 0 {
		return fmt.Errorf("invalid epoch for snapshot put")
	}

	return stsm.PutInEpochWithoutCache(key, val, stsm.epoch-1)
}

// GetFromOldEpochsWithoutAddingToCache searches for the key without adding it to cache if it is found. It starts
// the search for the key from the currentStorer - epochOffset)
func (stsm *snapshotTrieStorageManager) GetFromOldEpochsWithoutAddingToCache(key []byte, epochOffset int) ([]byte, error) {
	stsm.tsm.storageOperationMutex.Lock()
	defer stsm.tsm.storageOperationMutex.Unlock()

	if stsm.tsm.closed {
		log.Debug("getFromOldEpochsWithoutAddingToCache context closing, key = %s", hex.EncodeToString(key))
		return nil, errors.ErrContextClosing
	}

	return stsm.pruningStorer.GetFromOldEpochsWithoutAddingToCache(key, epochOffset)
}

// PutInEpochWithoutCache saves the data in the given epoch storer
func (stsm *snapshotTrieStorageManager) PutInEpochWithoutCache(key []byte, data []byte, epoch uint32) error {
	stsm.tsm.storageOperationMutex.Lock()
	defer stsm.tsm.storageOperationMutex.Unlock()

	if stsm.tsm.closed {
		log.Debug("putInEpochWithoutCache context closing, key = %s", hex.EncodeToString(key))
		return errors.ErrContextClosing
	}

	return stsm.pruningStorer.PutInEpochWithoutCache(key, data, epoch)
}

// GetFromEpochWithoutCache seeks the key in the specified epoch storer
func (stsm *snapshotTrieStorageManager) GetFromEpochWithoutCache(key []byte, epoch uint32) ([]byte, error) {
	stsm.tsm.storageOperationMutex.Lock()
	defer stsm.tsm.storageOperationMutex.Unlock()

	if stsm.tsm.closed {
		log.Debug("getFromEpochWithoutCache context closing, key = %s", hex.EncodeToString(key))
		return nil, errors.ErrContextClosing
	}

	return stsm.pruningStorer.GetFromEpochWithoutCache(key, epoch)
}

// RemoveFromCurrentEpoch removes the key from the current epoch storer
func (stsm *snapshotTrieStorageManager) RemoveFromCurrentEpoch(key []byte) error {
	stsm.tsm.storageOperationMutex.Lock()
	defer stsm.tsm.storageOperationMutex.Unlock()

	if stsm.tsm.closed {
		log.Debug("removeFromCurrentEpoch context closing, key = %s", hex.EncodeToString(key))
		return errors.ErrContextClosing
	}

	return stsm.pruningStorer.RemoveFromCurrentEpoch(key)
}

// GetLatestStorageEpoch returns the latest storage epoch
func (stsm *snapshotTrieStorageManager) GetLatestStorageEpoch() (uint32, error) {
	return stsm.epoch, nil
}
