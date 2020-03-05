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
	NodeDisplayName string
	PeerType        string
}

// PubKeyHeartbeat returns the heartbeat status for a public key
type PubKeyHeartbeat struct {
	TimeStamp       time.Time `json:"timeStamp"`
	HexPublicKey    string    `json:"hexPublicKey"`
	VersionNumber   string    `json:"versionNumber"`
	NodeDisplayName string    `json:"nodeDisplayName"`
	TotalUpTime     int       `json:"totalUpTimeSec"`
	TotalDownTime   int       `json:"totalDownTimeSec"`
	MaxInactiveTime Duration  `json:"maxInactiveTime"`
	ReceivedShardID uint32    `json:"receivedShardID"`
	ComputedShardID uint32    `json:"computedShardID"`
	PeerType        string    `json:"peerType"`
	IsActive        bool      `json:"isActive"`
}

// HeartbeatDTO is the struct used for handling DB operations for heartbeatMessageInfo struct
type HeartbeatDTO struct {
	LastUptimeDowntime          time.Time
	GenesisTime                 time.Time
	TimeStamp                   time.Time
	VersionNumber               string
	NodeDisplayName             string
	PeerType                    string
	MaxDurationPeerUnresponsive time.Duration
	MaxInactiveTime             Duration
	TotalUpTime                 Duration
	TotalDownTime               Duration
	ReceivedShardID             uint32
	ComputedShardID             uint32
	IsActive                    bool
}
