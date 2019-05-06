package heartbeat

import "github.com/ElrondNetwork/elrond-go-sandbox/p2p"

// PeerMessenger defines a subset of the p2p.Messenger interface
type PeerMessenger interface {
	Broadcast(topic string, buff []byte)
	PeerAddress(pid p2p.PeerID) string
}
