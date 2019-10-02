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
}

// PubKeyHeartbeat returns the heartbeat status for a public key
type PubKeyHeartbeat struct {
	HexPublicKey    string    `json:"hexPublicKey"`
	TimeStamp       time.Time `json:"timeStamp"`
	MaxInactiveTime Duration  `json:"maxInactiveTime"`
	IsActive        bool      `json:"isActive"`
	ReceivedShardID uint32    `json:"receivedShardID"`
	ComputedShardID uint32    `json:"computedShardID"`
	TotalUpTime     int       `json:"totalUpTimeSec"`
	TotalDownTime   int       `json:"totalDownTimeSec"`
	VersionNumber   string    `json:"versionNumber"`
	IsValidator     bool      `json:"isValidator"`
	NodeDisplayName string    `json:"nodeDisplayName"`
}

// HeartbeatDTO is the struct used for handling DB operations for heartbeatMessageInfo struct
type HeartbeatDTO struct {
	MaxDurationPeerUnresponsive time.Duration
	MaxInactiveTime             Duration
	TotalUpTime                 Duration
	TotalDownTime               Duration
	TimeStamp                   time.Time
	IsActive                    bool
	ReceivedShardID             uint32
	ComputedShardID             uint32
	VersionNumber               string
	NodeDisplayName             string
	IsValidator                 bool
	LastUptimeDowntime          time.Time
	GenesisTime                 time.Time
}
