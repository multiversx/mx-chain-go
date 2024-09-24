package receiptslog

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/storage"
)

type storageManagerOnlyGet struct {
	db storage.Persister
}

func NewStorageManagerOnlyGet(db storage.Persister) (*storageManagerOnlyGet, error) {
	return &storageManagerOnlyGet{
		db: db,
	}, nil
}

// Put -
func (s storageManagerOnlyGet) Put(_, _ []byte) error {
	return nil
}

// Get will return the data from local db based on the provided key
func (s storageManagerOnlyGet) Get(key []byte) ([]byte, error) {
	return s.db.Get(key)
}

// Remove -
func (s storageManagerOnlyGet) Remove(_ []byte) error {
	return nil
}

// Close -
func (s storageManagerOnlyGet) Close() error {
	return nil
}

// IsInterfaceNil -
func (s storageManagerOnlyGet) IsInterfaceNil() bool {
	return false
}

// PutInEpoch -
func (s storageManagerOnlyGet) PutInEpoch(_ []byte, _ []byte, _ uint32) error {
	return nil
}

// GetIdentifier -
func (s storageManagerOnlyGet) GetIdentifier() string {
	return ""
}

// GetStateStatsHandler -
func (s storageManagerOnlyGet) GetStateStatsHandler() common.StateStatisticsHandler {
	return disabled.NewStateStatistics()
}

// GetFromCurrentEpoch -
func (s storageManagerOnlyGet) GetFromCurrentEpoch(_ []byte) ([]byte, error) {
	return nil, nil
}

func (s storageManagerOnlyGet) PutInEpochWithoutCache(_ []byte, _ []byte, _ uint32) error {
	return nil
}

// TakeSnapshot -
func (s storageManagerOnlyGet) TakeSnapshot(_ string, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
}

// GetLatestStorageEpoch -
func (s storageManagerOnlyGet) GetLatestStorageEpoch() (uint32, error) {
	return 0, nil
}

// IsPruningEnabled -
func (s storageManagerOnlyGet) IsPruningEnabled() bool {
	return false
}

// IsPruningBlocked -
func (s storageManagerOnlyGet) IsPruningBlocked() bool {
	return false
}

// EnterPruningBufferingMode -
func (s storageManagerOnlyGet) EnterPruningBufferingMode() {
}

// ExitPruningBufferingMode -
func (s storageManagerOnlyGet) ExitPruningBufferingMode() {
}

// RemoveFromAllActiveEpochs -
func (s storageManagerOnlyGet) RemoveFromAllActiveEpochs(_ []byte) error {
	return nil
}

func (s storageManagerOnlyGet) SetEpochForPutOperation(_ uint32) {
}

// ShouldTakeSnapshot -
func (s storageManagerOnlyGet) ShouldTakeSnapshot() bool {
	return false
}

// IsSnapshotSupported -
func (s storageManagerOnlyGet) IsSnapshotSupported() bool {
	return false
}

// GetBaseTrieStorageManager -
func (s storageManagerOnlyGet) GetBaseTrieStorageManager() common.StorageManager {
	return nil
}

// IsClosed -
func (s storageManagerOnlyGet) IsClosed() bool {
	return false
}
