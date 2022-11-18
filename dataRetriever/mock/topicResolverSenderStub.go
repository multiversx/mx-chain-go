package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// TopicResolverSenderStub -
type TopicResolverSenderStub struct {
	SendCalled          func(buff []byte, peer core.PeerID) error
	TargetShardIDCalled func() uint32
	debugHandler        dataRetriever.ResolverDebugHandler
}

// RequestTopic -
func (trss *TopicResolverSenderStub) RequestTopic() string {
	return "topic_REQUEST"
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
