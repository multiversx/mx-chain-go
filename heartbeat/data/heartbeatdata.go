package data

import (
	"time"
)

// PubKeyHeartbeat returns the heartbeat status for a public key
type PubKeyHeartbeat struct {
	PublicKey            string    `json:"publicKey"`
	TimeStamp            time.Time `json:"timeStamp"`
	IsActive             bool      `json:"isActive"`
	ReceivedShardID      uint32    `json:"receivedShardID"`
	ComputedShardID      uint32    `json:"computedShardID"`
	VersionNumber        string    `json:"versionNumber"`
	NodeDisplayName      string    `json:"nodeDisplayName"`
	Identity             string    `json:"identity"`
	PeerType             string    `json:"peerType"`
	Nonce                uint64    `json:"nonce"`
	NumInstances         uint64    `json:"numInstances"`
	PeerSubType          uint32    `json:"peerSubType"`
	PidString            string    `json:"pidString"`
	NumTrieNodesReceived uint64    `json:"numTrieNodesReceived,omitempty"`
}
