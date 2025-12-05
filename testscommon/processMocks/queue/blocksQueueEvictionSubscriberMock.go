package queue

// BlocksQueueEvictionSubscriberMock -
type BlocksQueueEvictionSubscriberMock struct {
	OnHeaderEvictedCalled func(headerNonce uint64)
}

// OnHeaderEvicted -
func (mock *BlocksQueueEvictionSubscriberMock) OnHeaderEvicted(headerNonce uint64) {
	if mock.OnHeaderEvictedCalled != nil {
		mock.OnHeaderEvictedCalled(headerNonce)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mock *BlocksQueueEvictionSubscriberMock) IsInterfaceNil() bool {
	return mock == nil
}
