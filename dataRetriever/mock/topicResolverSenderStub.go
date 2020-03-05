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

// Topic -
func (trss *TopicResolverSenderStub) Topic() string {
	return "_REQUEST"
}

// SendOnRequestTopic -
func (trss *TopicResolverSenderStub) SendOnRequestTopic(rd *dataRetriever.RequestData) error {
	if trss.SendOnRequestTopicCalled != nil {
		return trss.SendOnRequestTopicCalled(rd)
	}

	return nil
}

// Send -
func (trss *TopicResolverSenderStub) Send(buff []byte, peer p2p.PeerID) error {
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

// IsInterfaceNil returns true if there is no value under the interface
func (trss *TopicResolverSenderStub) IsInterfaceNil() bool {
	return trss == nil
}
