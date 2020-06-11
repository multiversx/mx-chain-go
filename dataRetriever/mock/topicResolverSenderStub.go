package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// TopicResolverSenderStub -
type TopicResolverSenderStub struct {
	SendOnRequestTopicCalled func(rd *dataRetriever.RequestData, originalHashes [][]byte) error
	SendCalled               func(buff []byte, peer core.PeerID) error
	TargetShardIDCalled      func() uint32
	SetNumPeersToQueryCalled func(intra int, cross int)
	GetNumPeersToQueryCalled func() (int, int)
	debugHandler             dataRetriever.ResolverDebugHandler
}

// SetNumPeersToQuery -
func (trss *TopicResolverSenderStub) SetNumPeersToQuery(intra int, cross int) {
	if trss.SetNumPeersToQueryCalled != nil {
		trss.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (trss *TopicResolverSenderStub) NumPeersToQuery() (int, int) {
	if trss.GetNumPeersToQueryCalled != nil {
		return trss.GetNumPeersToQueryCalled()
	}

	return 2, 2
}

// RequestTopic -
func (trss *TopicResolverSenderStub) RequestTopic() string {
	return "topic_REQUEST"
}

// SendOnRequestTopic -
func (trss *TopicResolverSenderStub) SendOnRequestTopic(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
	if trss.SendOnRequestTopicCalled != nil {
		return trss.SendOnRequestTopicCalled(rd, originalHashes)
	}

	return nil
}

// Send -
func (trss *TopicResolverSenderStub) Send(buff []byte, peer core.PeerID) error {
	if trss.SendCalled != nil {
		return trss.SendCalled(buff, peer)
	}

	return nil
}

// TargetShardID -
func (trss *TopicResolverSenderStub) TargetShardID() uint32 {
	if trss.TargetShardIDCalled != nil {
		return trss.TargetShardIDCalled()
	}

	return 0
}

// ResolverDebugHandler -
func (trss *TopicResolverSenderStub) ResolverDebugHandler() dataRetriever.ResolverDebugHandler {
	if check.IfNil(trss.debugHandler) {
		return &ResolverDebugHandler{}
	}

	return trss.debugHandler
}

// SetResolverDebugHandler -
func (trss *TopicResolverSenderStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	trss.debugHandler = handler

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (trss *TopicResolverSenderStub) IsInterfaceNil() bool {
	return trss == nil
}
