package heartbeat

import "github.com/ElrondNetwork/elrond-go-sandbox/p2p"

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	Broadcast(topic string, buff []byte)
	PeerAddress(pid p2p.PeerID) string
}
