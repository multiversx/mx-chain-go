package core

// P2PPeerType defines the type of a p2p peer
type P2PPeerType int

const (
	// ValidatorPeer means that the peer is a validator
	ValidatorPeer P2PPeerType = 1
	// ObserverdPeer means that the peer is an observer
	ObserverdPeer P2PPeerType = 2
	// UnknownPeer defines a peer that is unknown (did not advertise data in any way)
	UnknownPeer P2PPeerType = 3
)

// P2PPeerInfo represents a peer info structure
type P2PPeerInfo struct {
	PeerType P2PPeerType
	ShardID  uint32
}
