package containers

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ dataRetriever.ResolversFinder = (*resolversFinder)(nil)

// resolversFinder is an implementation of process.ResolverContainer meant to be used
// wherever a resolver fetch is required
type resolversFinder struct {
	dataRetriever.ResolversContainer
	coordinator sharding.Coordinator
}

// NewResolversFinder creates a new resolversFinder object
func NewResolversFinder(container dataRetriever.ResolversContainer, coordinator sharding.Coordinator) (*resolversFinder, error) {
	if check.IfNil(container) {
		return nil, dataRetriever.ErrNilResolverContainer
	}

	if check.IfNil(coordinator) {
		return nil, dataRetriever.ErrNilShardCoordinator
	}

	return &resolversFinder{
		ResolversContainer: container,
		coordinator:        coordinator,
	}, nil
}

// IntraShardResolver fetches the intrashard Resolver starting from a baseTopic
// baseTopic will be one of the constants defined in factory.go: TransactionTopic, HeadersTopic and so on
func (rf *resolversFinder) IntraShardResolver(baseTopic string) (dataRetriever.Resolver, error) {
	topic := baseTopic + rf.coordinator.CommunicationIdentifier(rf.coordinator.SelfId())
	return rf.Get(topic)
}

// MetaChainResolver fetches the metachain Resolver starting from a baseTopic
// baseTopic will be one of the constants defined in factory.go: metaHeaderTopic, MetaPeerChangeTopic and so on
func (rf *resolversFinder) MetaChainResolver(baseTopic string) (dataRetriever.Resolver, error) {
	return rf.Get(baseTopic)
}

// CrossShardResolver fetches the cross shard Resolver starting from a baseTopic and a cross shard id
// baseTopic will be one of the constants defined in factory.go: TransactionTopic, HeadersTopic and so on
func (rf *resolversFinder) CrossShardResolver(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error) {
	topic := baseTopic + rf.coordinator.CommunicationIdentifier(crossShard)
	return rf.Get(topic)
}

// MetaCrossShardResolver fetches the cross shard Resolver between crossShard and meta
func (rf *resolversFinder) MetaCrossShardResolver(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error) {
	topic := baseTopic + core.CommunicationIdentifierBetweenShards(crossShard, core.MetachainShardId)
	return rf.Get(topic)
}

// IsInterfaceNil returns true if underlying struct is nil
func (rf *resolversFinder) IsInterfaceNil() bool {
	return rf == nil || check.IfNil(rf.ResolversContainer)
}
