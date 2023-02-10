package storage

// ShardIDProviderStub implements ShardIDProvider interface
type ShardIDProviderStub struct {
	ComputeIdCalled      func(key []byte) uint32
	NumberOfShardsCalled func() uint32
	GetShardIDsCalled    func() []uint32
}

// ComputeId -
func (stub *ShardIDProviderStub) ComputeId(key []byte) uint32 {
	if stub.ComputeIdCalled != nil {
		return stub.ComputeIdCalled(key)
	}

	return 0
}

// NumberOfShards -
func (stub *ShardIDProviderStub) NumberOfShards() uint32 {
	if stub.NumberOfShardsCalled != nil {
		return stub.NumberOfShardsCalled()
	}

	return 0
}

// GetShardIDs -
func (stub *ShardIDProviderStub) GetShardIDs() []uint32 {
	if stub.GetShardIDsCalled != nil {
		return stub.GetShardIDsCalled()
	}

	return make([]uint32, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *ShardIDProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
