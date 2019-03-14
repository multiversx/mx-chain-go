package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type TopicResolverSenderStub struct {
	SendOnRequestTopicCalled func(rd *process.RequestData) error
	SendCalled               func(buff []byte, peer p2p.PeerID) error
}

func (trss *TopicResolverSenderStub) TopicRequestSuffix() string {
	return "_REQUEST"
}

func (trss *TopicResolverSenderStub) SendOnRequestTopic(rd *process.RequestData) error {
	return trss.SendOnRequestTopicCalled(rd)
}

func (trss *TopicResolverSenderStub) Send(buff []byte, peer p2p.PeerID) error {
	return trss.SendCalled(buff, peer)
}
