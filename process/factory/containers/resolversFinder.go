package containers

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// resolversFinder is an implementation of process.ResolverContainer meant to be used
// wherever a resolver fetch is required
type resolversFinder struct {
	process.ResolversContainer
	coordinator sharding.Coordinator
}

// NewResolversFinder creates a new resolversFinder object
func NewResolversFinder(container process.ResolversContainer, coordinator sharding.Coordinator) (*resolversFinder, error) {
	if container == nil {
		return nil, process.ErrNilResolverContainer
	}

	if coordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	return &resolversFinder{
		ResolversContainer: container,
		coordinator:        coordinator,
	}, nil
}

// IntraShardResolver fetches the intrashard Resolver starting from a baseTopic
// baseTopic will be one of the constants defined in factory.go: TransactionTopic, HeadersTopic and so on
func (rf *resolversFinder) IntraShardResolver(baseTopic string) (process.Resolver, error) {
	topic := baseTopic + rf.coordinator.CommunicationIdentifier(rf.coordinator.SelfId())

	return rf.Get(topic)
}

// CrossShardResolver fetches the crossshard Resolver starting from a baseTopic and a cross shard id
// baseTopic will be one of the constants defined in factory.go: TransactionTopic, HeadersTopic and so on
func (rf *resolversFinder) CrossShardResolver(baseTopic string, crossShard uint32) (process.Resolver, error) {
	topic := baseTopic + rf.coordinator.CommunicationIdentifier(crossShard)

	return rf.Get(topic)
}
