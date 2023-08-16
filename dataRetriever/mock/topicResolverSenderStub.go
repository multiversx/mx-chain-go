package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

// TopicResolverSenderStub -
type TopicResolverSenderStub struct {
	SendCalled          func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error
	TargetShardIDCalled func() uint32
	debugHandler        dataRetriever.DebugHandler
}

// RequestTopic -
func (trss *TopicResolverSenderStub) RequestTopic() string {
	return "topic_REQUEST"
}

// Send -
func (trss *TopicResolverSenderStub) Send(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
	if trss.SendCalled != nil {
		return trss.SendCalled(buff, peer, source)
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

// DebugHandler -
func (trss *TopicResolverSenderStub) DebugHandler() dataRetriever.DebugHandler {
	if check.IfNil(trss.debugHandler) {
		return &DebugHandler{}
	}

	return trss.debugHandler
}

// SetDebugHandler -
func (trss *TopicResolverSenderStub) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	trss.debugHandler = handler

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (trss *TopicResolverSenderStub) IsInterfaceNil() bool {
	return trss == nil
}
