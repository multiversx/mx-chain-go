package heartbeat

import (
	"time"
)

// Heartbeat represents the heartbeat message that is sent between peers
type Heartbeat struct {
	Payload         []byte
	Pubkey          []byte
	Signature       []byte
	ShardID         uint32
	VersionNumber   string
	IsValidator     bool
	NodeDisplayName string
}

// PubKeyHeartbeat returns the heartbeat status for the public key
type PubKeyHeartbeat struct {
	HexPublicKey    string
	TimeStamp       time.Time
	MaxInactiveTime Duration
	IsActive        bool
	ShardID         uint32
	TotalUpTime     Duration
	TotalDownTime   Duration
	VersionNumber   string
	IsValidator     bool
	NodeDisplayName string
}
