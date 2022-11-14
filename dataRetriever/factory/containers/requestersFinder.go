package containers

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ dataRetriever.RequestersFinder = (*requestersFinder)(nil)

// requestersFinder is an implementation of process.RequesterContainer meant to be used
// wherever a resolver fetch is required
type requestersFinder struct {
	dataRetriever.RequestersContainer
	coordinator sharding.Coordinator
}

// NewRequestersFinder creates a new requestersFinder object
func NewRequestersFinder(container dataRetriever.RequestersContainer, coordinator sharding.Coordinator) (*requestersFinder, error) {
	if check.IfNil(container) {
		return nil, dataRetriever.ErrNilRequestersContainer
	}

	if check.IfNil(coordinator) {
		return nil, dataRetriever.ErrNilShardCoordinator
	}

	return &requestersFinder{
		RequestersContainer: container,
		coordinator:         coordinator,
	}, nil
}

// IntraShardRequester fetches the intrashard Requester starting from a baseTopic
// baseTopic will be one of the constants defined in factory.go: TransactionTopic, HeadersTopic and so on
func (rf *requestersFinder) IntraShardRequester(baseTopic string) (dataRetriever.Requester, error) {
	topic := baseTopic + rf.coordinator.CommunicationIdentifier(rf.coordinator.SelfId())
	return rf.Get(topic)
}

// MetaChainRequester fetches the metachain Requester starting from a baseTopic
// baseTopic will be one of the constants defined in factory.go: metaHeaderTopic, MetaPeerChangeTopic and so on
func (rf *requestersFinder) MetaChainRequester(baseTopic string) (dataRetriever.Requester, error) {
	return rf.Get(baseTopic)
}

// CrossShardRequester fetches the cross shard Requester starting from a baseTopic and a cross shard id
// baseTopic will be one of the constants defined in factory.go: TransactionTopic, HeadersTopic and so on
func (rf *requestersFinder) CrossShardRequester(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
	topic := baseTopic + rf.coordinator.CommunicationIdentifier(crossShard)
	return rf.Get(topic)
}

// MetaCrossShardRequester fetches the cross shard Requester between crossShard and meta
func (rf *requestersFinder) MetaCrossShardRequester(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
	topic := baseTopic + core.CommunicationIdentifierBetweenShards(crossShard, core.MetachainShardId)
	return rf.Get(topic)
}

// IsInterfaceNil returns true if underlying struct is nil
func (rf *requestersFinder) IsInterfaceNil() bool {
	return rf == nil || check.IfNil(rf.RequestersContainer)
}
