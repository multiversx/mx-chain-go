package queue

// BlocksQueueEvictionSubscriber defines a component that will be notified when a header is removed from the processing queue
type BlocksQueueEvictionSubscriber interface {
	OnHeaderEvicted(headerNonce uint64)
	IsInterfaceNil() bool
}
