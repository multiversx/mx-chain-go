package trie

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type trieStorageManagerWithoutSnapshot struct {
	common.StorageManager
}

// NewTrieStorageManagerWithoutSnapshot creates a new trieStorageManagerWithoutSnapshot
func NewTrieStorageManagerWithoutSnapshot(tsm common.StorageManager) (*trieStorageManagerWithoutSnapshot, error) {
	if check.IfNil(tsm) {
		return nil, ErrNilTrieStorage
	}

	return &trieStorageManagerWithoutSnapshot{
		StorageManager: tsm,
	}, nil
}

// GetFromCurrentEpoch calls Get(), as this implementation uses a static storer
func (tsm *trieStorageManagerWithoutSnapshot) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	return tsm.Get(key)
}

// PutInEpoch calls Put(), as this implementation uses a static storer
func (tsm *trieStorageManagerWithoutSnapshot) PutInEpoch(key []byte, val []byte, _ uint32) error {
	return tsm.Put(key, val)
}

// PutInEpochWithoutCache calls Put(), as this implementation uses a static storer
func (tsm *trieStorageManagerWithoutSnapshot) PutInEpochWithoutCache(key []byte, val []byte, _ uint32) error {
	return tsm.Put(key, val)
}

// TakeSnapshot does nothing, as snapshots are disabled for this implementation
func (tsm *trieStorageManagerWithoutSnapshot) TakeSnapshot(_ string, _ []byte, _ []byte, iteratorChannels *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
	if iteratorChannels != nil {
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
	}
	stats.SnapshotFinished()
}

// GetLatestStorageEpoch returns 0, as this implementation uses a static storer
func (tsm *trieStorageManagerWithoutSnapshot) GetLatestStorageEpoch() (uint32, error) {
	return 0, nil
}

// SetEpochForPutOperation does nothing for this implementation
func (tsm *trieStorageManagerWithoutSnapshot) SetEpochForPutOperation(uint32) {
}

// ShouldTakeSnapshot returns false for this implementation
func (tsm *trieStorageManagerWithoutSnapshot) ShouldTakeSnapshot() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (tsm *trieStorageManagerWithoutSnapshot) IsInterfaceNil() bool {
	return tsm == nil
}
