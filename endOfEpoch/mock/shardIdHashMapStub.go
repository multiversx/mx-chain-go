package mock

type ShardIdHasMapStub struct {
	LoadCalled   func(shardId uint32) ([]byte, bool)
	StoreCalled  func(shardId uint32, hash []byte)
	RangeCalled  func(f func(shardId uint32, hash []byte) bool)
	DeleteCalled func(shardId uint32)
}

func (sihsm *ShardIdHasMapStub) Load(shardId uint32) ([]byte, bool) {
	return sihsm.LoadCalled(shardId)
}

func (sihsm *ShardIdHasMapStub) Store(shardId uint32, hash []byte) {
	sihsm.StoreCalled(shardId, hash)
}

func (sihsm *ShardIdHasMapStub) Range(f func(shardId uint32, hash []byte) bool) {
	sihsm.RangeCalled(f)
}

func (sihsm *ShardIdHasMapStub) Delete(shardId uint32) {
	sihsm.DeleteCalled(shardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sihsm *ShardIdHasMapStub) IsInterfaceNil() bool {
	if sihsm == nil {
		return true
	}
	return false
}
