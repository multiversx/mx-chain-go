package trie

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
)

type syncTrieStorageManager struct {
	common.StorageManagerWithStats
	epoch uint32
}

// NewSyncTrieStorageManager creates a new instance of syncTrieStorageManager
func NewSyncTrieStorageManager(tsm common.StorageManager) (*syncTrieStorageManager, error) {
	if check.IfNil(tsm) {
		return nil, ErrNilTrieStorage
	}

	epoch, err := tsm.GetLatestStorageEpoch()
	if err != nil {
		return nil, err
	}

	tsmWithStats, ok := tsm.(common.StorageManagerWithStats)
	if !ok {
		return nil, errors.ErrWrongTypeAssertion
	}

	return &syncTrieStorageManager{
		StorageManagerWithStats: tsmWithStats,
		epoch:                   epoch,
	}, nil
}

// Put adds the given value to the current and previous epoch storer. This is done only when syncing.
func (stsm *syncTrieStorageManager) Put(key []byte, val []byte) error {
	err := stsm.PutInEpoch(key, val, stsm.epoch)
	if err != nil {
		return err
	}

	if stsm.epoch == 0 {
		return nil
	}

	return stsm.PutInEpoch(key, val, stsm.epoch-1)
}
