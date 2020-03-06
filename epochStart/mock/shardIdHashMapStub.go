package mock

// ShardIdHasMapStub -
type ShardIdHasMapStub struct {
	LoadCalled   func(shardId uint32) ([]byte, bool)
	StoreCalled  func(shardId uint32, hash []byte)
	RangeCalled  func(f func(shardId uint32, hash []byte) bool)
	DeleteCalled func(shardId uint32)
}

// Load -
func (sihsm *ShardIdHasMapStub) Load(shardId uint32) ([]byte, bool) {
	return sihsm.LoadCalled(shardId)
}

// Store -
func (sihsm *ShardIdHasMapStub) Store(shardId uint32, hash []byte) {
	sihsm.StoreCalled(shardId, hash)
}

// Range -
func (sihsm *ShardIdHasMapStub) Range(f func(shardId uint32, hash []byte) bool) {
	sihsm.RangeCalled(f)
}

// Delete -
func (sihsm *ShardIdHasMapStub) Delete(shardId uint32) {
	sihsm.DeleteCalled(shardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sihsm *ShardIdHasMapStub) IsInterfaceNil() bool {
	return sihsm == nil
}
