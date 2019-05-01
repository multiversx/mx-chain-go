package heartbeat

import (
	"time"
)

// Heartbeat represents the heartbeat message that is sent between peers
type Heartbeat struct {
	Payload   []byte
	Pubkey    []byte
	Signature []byte
}

// PeerHeartbeat represents the status of a received message from a p2p address
type PeerHeartbeat struct {
	P2PAddress      string
	TimeStamp       time.Time
	MaxInactiveTime time.Duration
	IsActive        bool
}

// HeartbeatStatus returns the heartbeat status for each public key
type HeartbeatStatus struct {
	HexPublicKey   string
	PeerHeartBeats []PeerHeartbeat
}
