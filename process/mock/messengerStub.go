package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MessengerStub struct {
	CloseCalled                      func() error
	IDCalled                         func() p2p.PeerID
	PeersCalled                      func() []p2p.PeerID
	AddressesCalled                  func() []string
	ConnectToPeerCalled              func(address string) error
	KadDhtDiscoverNewPeersCalled     func() error
	TrimConnectionsCalled            func()
	IsConnectedCalled                func(peerID p2p.PeerID) bool
	ConnectedPeersCalled             func() []p2p.PeerID
	CreateTopicCalled                func(name string, createPipeForTopic bool) error
	HasTopicCalled                   func(name string) bool
	HasTopicValidatorCalled          func(name string) bool
	BroadcastOnPipeCalled            func(pipe string, topic string, buff []byte)
	BroadcastCalled                  func(topic string, buff []byte)
	RegisterMessageProcessorCalled   func(topic string, handler p2p.MessageProcessor) error
	UnregisterMessageProcessorCalled func(topic string) error
	SendToConnectedPeerCalled        func(topic string, buff []byte, peerID p2p.PeerID) error
	OutgoingPipeLoadBalancerCalled   func() p2p.PipeLoadBalancer
	BootstrapCalled                  func() error
}

func (ms *MessengerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	return ms.RegisterMessageProcessorCalled(topic, handler)
}

func (ms *MessengerStub) UnregisterMessageProcessor(topic string) error {
	return ms.UnregisterMessageProcessorCalled(topic)
}

func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

func (ms *MessengerStub) OutgoingPipeLoadBalancer() p2p.PipeLoadBalancer {
	return ms.OutgoingPipeLoadBalancerCalled()
}

func (ms *MessengerStub) Close() error {
	return ms.CloseCalled()
}

func (ms *MessengerStub) ID() p2p.PeerID {
	return ms.IDCalled()
}

func (ms *MessengerStub) Peers() []p2p.PeerID {
	return ms.PeersCalled()
}

func (ms *MessengerStub) Addresses() []string {
	return ms.AddressesCalled()
}

func (ms *MessengerStub) ConnectToPeer(address string) error {
	return ms.ConnectToPeerCalled(address)
}

func (ms *MessengerStub) KadDhtDiscoverNewPeers() error {
	return ms.KadDhtDiscoverNewPeersCalled()
}

func (ms *MessengerStub) TrimConnections() {
	ms.TrimConnectionsCalled()
}

func (ms *MessengerStub) IsConnected(peerID p2p.PeerID) bool {
	return ms.IsConnectedCalled(peerID)
}

func (ms *MessengerStub) ConnectedPeers() []p2p.PeerID {
	return ms.ConnectedPeersCalled()
}

func (ms *MessengerStub) CreateTopic(name string, createPipeForTopic bool) error {
	return ms.CreateTopicCalled(name, createPipeForTopic)
}

func (ms *MessengerStub) HasTopic(name string) bool {
	return ms.HasTopicCalled(name)
}

func (ms *MessengerStub) HasTopicValidator(name string) bool {
	return ms.HasTopicValidatorCalled(name)
}

func (ms *MessengerStub) BroadcastOnPipe(pipe string, topic string, buff []byte) {
	ms.BroadcastOnPipeCalled(pipe, topic, buff)
}

func (ms *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return ms.SendToConnectedPeerCalled(topic, buff, peerID)
}

func (ms *MessengerStub) Bootstrap() error {
	return ms.BootstrapCalled()
}
