//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. heartbeat.proto
package data

import (
	"encoding/json"
	"errors"
	"time"
)

// PubKeyHeartbeat returns the heartbeat status for a public key
type PubKeyHeartbeat struct {
	PublicKey       string    `json:"publicKey"`
	TimeStamp       time.Time `json:"timeStamp"`
	MaxInactiveTime Duration  `json:"maxInactiveTime"`
	IsActive        bool      `json:"isActive"`
	ReceivedShardID uint32    `json:"receivedShardID"`
	ComputedShardID uint32    `json:"computedShardID"`
	TotalUpTime     int64     `json:"totalUpTimeSec"`
	TotalDownTime   int64     `json:"totalDownTimeSec"`
	VersionNumber   string    `json:"versionNumber"`
	NodeDisplayName string    `json:"nodeDisplayName"`
	Identity        string    `json:"identity"`
	PeerType        string    `json:"peerType"`
	Nonce           uint64    `json:"nonce"`
	NumInstances    uint64    `json:"numInstances"`
	PeerSubType     uint32    `json:"peerSubType"`
}

// Duration is a wrapper of the original Duration struct
// that has JSON marshal and unmarshal capabilities
// golang issue: https://github.com/golang/go/issues/10275
type Duration struct {
	time.Duration
}

// MarshalJSON is called when a json marshal is triggered on this field
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON is called when a json unmarshal is triggered on this field
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
