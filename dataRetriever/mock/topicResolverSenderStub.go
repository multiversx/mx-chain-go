package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// TopicResolverSenderStub -
type TopicResolverSenderStub struct {
	SendOnRequestTopicCalled func(rd *dataRetriever.RequestData) error
	SendCalled               func(buff []byte, peer p2p.PeerID) error
	TargetShardIDCalled      func() uint32
}

// TopicRequestSuffix -
func (trss *TopicResolverSenderStub) TopicRequestSuffix() string {
	return "_REQUEST"
}

// SendOnRequestTopic -
func (trss *TopicResolverSenderStub) SendOnRequestTopic(rd *dataRetriever.RequestData) error {
	return trss.SendOnRequestTopicCalled(rd)
}

// Send -
func (trss *TopicResolverSenderStub) Send(buff []byte, peer p2p.PeerID) error {
	return trss.SendCalled(buff, peer)
}

// TargetShardID -
func (trss *TopicResolverSenderStub) TargetShardID() uint32 {
	return trss.TargetShardIDCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (trss *TopicResolverSenderStub) IsInterfaceNil() bool {
	return trss == nil
}
