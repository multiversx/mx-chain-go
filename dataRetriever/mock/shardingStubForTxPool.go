package mock

type ShardingStubForTxPool struct {
	numShards uint32
}

func NewShardingStubForTxPool(numShards uint32) *ShardingStubForTxPool {
	return &ShardingStubForTxPool{
		numShards: numShards,
	}
}

func (stub *ShardingStubForTxPool) NumberOfShards() uint32 {
	return stub.numShards
}

func (stub *ShardingStubForTxPool) IsInterfaceNil() bool {
	return stub == nil
}
