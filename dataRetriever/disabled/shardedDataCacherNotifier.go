package disabled

import (
	"github.com/multiversx/mx-chain-core-go/core/counting"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/disabled"
)

type shardedDataCacherNotifier struct {
}

// NewShardedDataCacherNotifier returns a new instance of disabled shardedDataCacherNotifier
func NewShardedDataCacherNotifier() *shardedDataCacherNotifier {
	return &shardedDataCacherNotifier{}
}

// RegisterOnAdded does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) RegisterOnAdded(_ func(key []byte, value interface{})) {
}

// ShardDataStore returns a disabled cache
func (sdcn *shardedDataCacherNotifier) ShardDataStore(_ string) (c storage.Cacher) {
	return disabled.NewCache()
}

// AddData does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) AddData(_ []byte, _ interface{}, _ int, _ string) {
}

// SearchFirstData returns false
func (sdcn *shardedDataCacherNotifier) SearchFirstData(_ []byte) (value interface{}, ok bool) {
	return nil, false
}

// RemoveData does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) RemoveData(_ []byte, _ string) {
}

// RemoveSetOfDataFromPool does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) RemoveSetOfDataFromPool(_ [][]byte, _ string) {
}

// ImmunizeSetOfDataAgainstEviction does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) ImmunizeSetOfDataAgainstEviction(_ [][]byte, _ string) {
}

// RemoveDataFromAllShards does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) RemoveDataFromAllShards(_ []byte) {
}

// MergeShardStores does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) MergeShardStores(_, _ string) {
}

// Clear does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) Clear() {
}

// ClearShardStore does nothing as it is disabled
func (sdcn *shardedDataCacherNotifier) ClearShardStore(_ string) {
}

// GetCounts returns a null counts
func (sdcn *shardedDataCacherNotifier) GetCounts() counting.CountsWithSize {
	return &counting.NullCounts{}
}

// Keys returns an empty slice
func (sdcn *shardedDataCacherNotifier) Keys() [][]byte {
	return make([][]byte, 0)
}

// Close returns nil
func (sdcn *shardedDataCacherNotifier) Close() error {
	return nil
}

// IsInterfaceNil return true if there is no value under the interface
func (sdcn *shardedDataCacherNotifier) IsInterfaceNil() bool {
	return sdcn == nil
}
