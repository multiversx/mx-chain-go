package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/p2p"

type P2PMessenger struct {
	BroadcastCalled   func(topic string, buff []byte)
	PeerAddressCalled func(pid p2p.PeerID) string
}

func (messenger *P2PMessenger) Broadcast(topic string, buff []byte) {
	messenger.BroadcastCalled(topic, buff)
}

func (messenger *P2PMessenger) PeerAddress(pid p2p.PeerID) string {
	return messenger.PeerAddressCalled(pid)
}
