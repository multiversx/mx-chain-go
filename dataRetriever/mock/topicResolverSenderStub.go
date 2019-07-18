package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type TopicResolverSenderStub struct {
	SendOnRequestTopicCalled func(rd *dataRetriever.RequestData) error
	SendCalled               func(buff []byte, peer p2p.PeerID) error
	TargetShardIDCalled      func() uint32
}

func (trss *TopicResolverSenderStub) TopicRequestSuffix() string {
	return "_REQUEST"
}

func (trss *TopicResolverSenderStub) SendOnRequestTopic(rd *dataRetriever.RequestData) error {
	return trss.SendOnRequestTopicCalled(rd)
}

func (trss *TopicResolverSenderStub) Send(buff []byte, peer p2p.PeerID) error {
	return trss.SendCalled(buff, peer)
}

func (trss *TopicResolverSenderStub) TargetShardID() uint32 {
	return trss.TargetShardIDCalled()
}
