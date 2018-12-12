package mock

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

type MessengerStub struct {
	marshalizer    marshal.Marshalizer
	HasherObj      hashing.Hasher
	AddTopicCalled func(t *p2p.Topic) error
	GetTopicCalled func(name string) *p2p.Topic
}

func NewMessengerStub() *MessengerStub {
	return &MessengerStub{
		marshalizer: &MarshalizerMock{},
		HasherObj:   HasherMock{},
	}
}

func (ms *MessengerStub) Close() error {
	panic("implement me")
}

func (ms *MessengerStub) ID() peer.ID {
	panic("implement me")
}

func (ms *MessengerStub) Peers() []peer.ID {
	panic("implement me")
}

func (ms *MessengerStub) Conns() []net.Conn {
	panic("implement me")
}

func (ms *MessengerStub) Marshalizer() marshal.Marshalizer {
	return ms.marshalizer
}

func (ms *MessengerStub) Hasher() hashing.Hasher {
	return ms.HasherObj
}

func (ms *MessengerStub) RouteTable() *p2p.RoutingTable {
	panic("implement me")
}

func (ms *MessengerStub) Addresses() []string {
	panic("implement me")
}

func (ms *MessengerStub) ConnectToAddresses(ctx context.Context, addresses []string) {
	panic("implement me")
}

func (ms *MessengerStub) Bootstrap(ctx context.Context) {
	panic("implement me")
}

func (ms *MessengerStub) PrintConnected() {
	panic("implement me")
}

func (ms *MessengerStub) AddAddress(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	panic("implement me")
}

func (ms *MessengerStub) Connectedness(pid peer.ID) net.Connectedness {
	panic("implement me")
}

func (ms *MessengerStub) GetTopic(topicName string) *p2p.Topic {
	return ms.GetTopicCalled(topicName)
}

func (ms *MessengerStub) AddTopic(t *p2p.Topic) error {
	return ms.AddTopicCalled(t)
}
