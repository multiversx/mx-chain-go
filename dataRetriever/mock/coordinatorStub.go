package mock

// CoordinatorStub -
type CoordinatorStub struct {
	NumberOfShardsCalled          func() uint32
	ComputeIdCalled               func(address []byte) uint32
	SelfIdCalled                  func() uint32
	SameShardCalled               func(firstAddress, secondAddress []byte) bool
	CommunicationIdentifierCalled func(destShardID uint32) string
}

// NumberOfShards -
func (coordinator *CoordinatorStub) NumberOfShards() uint32 {
	return coordinator.NumberOfShardsCalled()
}

// ComputeId -
func (coordinator *CoordinatorStub) ComputeId(address []byte) uint32 {
	return coordinator.ComputeIdCalled(address)
}

// SelfId -
func (coordinator *CoordinatorStub) SelfId() uint32 {
	return coordinator.SelfIdCalled()
}

// SameShard -
func (coordinator *CoordinatorStub) SameShard(firstAddress, secondAddress []byte) bool {
	return coordinator.SameShardCalled(firstAddress, secondAddress)
}

// CommunicationIdentifier -
func (coordinator *CoordinatorStub) CommunicationIdentifier(destShardID uint32) string {
	return coordinator.CommunicationIdentifierCalled(destShardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (coordinator *CoordinatorStub) IsInterfaceNil() bool {
	return coordinator == nil
}
