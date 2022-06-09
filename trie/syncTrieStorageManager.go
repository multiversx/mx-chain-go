package trie

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

type syncTrieStorageManager struct {
	common.StorageManager
	epoch uint32
}

// NewSyncTrieStorageManager creates a new instance of syncTrieStorageManager
func NewSyncTrieStorageManager(tsm common.StorageManager) (*syncTrieStorageManager, error) {
	epoch, err := tsm.GetLatestStorageEpoch()
	if err != nil {
		return nil, err
	}

	return &syncTrieStorageManager{
		StorageManager: tsm,
		epoch:          epoch,
	}, nil
}

// Put adds the given value to the previous epoch storer if available, else it adds it to the current storer
func (stsm *syncTrieStorageManager) Put(key []byte, val []byte) error {
	if stsm.epoch == 0 {
		return stsm.PutInEpoch(key, val, stsm.epoch)
	}

	return stsm.PutInEpoch(key, val, stsm.epoch-1)
}
