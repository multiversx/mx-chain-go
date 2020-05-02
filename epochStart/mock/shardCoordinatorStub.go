package mock

// ShardCoordinatorStub -
type ShardCoordinatorStub struct {
	NumberOfShardsCalled          func() uint32
	ComputeIdCalled               func(address []byte) uint32
	SelfIdCalled                  func() uint32
	SameShardCalled               func(firstAddress, secondAddress []byte) bool
	CommunicationIdentifierCalled func(destShardID uint32) string
}

// NumberOfShards -
func (coordinator *ShardCoordinatorStub) NumberOfShards() uint32 {
	if coordinator.NumberOfShardsCalled != nil {
		return coordinator.NumberOfShardsCalled()
	}
	return 1
}

// ComputeId -
func (coordinator *ShardCoordinatorStub) ComputeId(address []byte) uint32 {
	return coordinator.ComputeIdCalled(address)
}

// SelfId -
func (coordinator *ShardCoordinatorStub) SelfId() uint32 {
	if coordinator.SelfIdCalled != nil {
		return coordinator.SelfIdCalled()
	}
	return 0
}

// SameShard -
func (coordinator *ShardCoordinatorStub) SameShard(firstAddress, secondAddress []byte) bool {
	return coordinator.SameShardCalled(firstAddress, secondAddress)
}

// CommunicationIdentifier -
func (coordinator *ShardCoordinatorStub) CommunicationIdentifier(destShardID uint32) string {
	return coordinator.CommunicationIdentifierCalled(destShardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (coordinator *ShardCoordinatorStub) IsInterfaceNil() bool {
	return coordinator == nil
}
