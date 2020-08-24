package process

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
)

const MaxSizeInBytes = maxSizeInBytes

// GetMessages -
func (m *Monitor) GetMessages() map[string]*heartbeatMessageInfo {
	return m.heartbeatMessages
}

// SetMessages -
func (m *Monitor) SetMessages(messages map[string]*heartbeatMessageInfo) {
	m.heartbeatMessages = messages
}

// GetHeartbeatMessageInfo -
func (m *Monitor) GetHeartbeatMessageInfo(_ time.Time) *heartbeatMessageInfo {
	return &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: 0,
		maxInactiveTime:             time.Duration(0),
		totalUpTime:                 time.Duration(0),
		totalDownTime:               time.Duration(0),
		getTimeHandler:              nil,
		timeStamp:                   time.Time{},
		isActive:                    false,
		receivedShardID:             0,
		computedShardID:             0,
		versionNumber:               "",
		nodeDisplayName:             "",
		lastUptimeDowntime:          time.Time{},
		genesisTime:                 time.Time{},
	}
}

// SendHeartbeatMessage -
func (m *Monitor) SendHeartbeatMessage(hb *data.Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

// AddHeartbeatMessageToMap -
func (m *Monitor) AddHeartbeatMessageToMap(hb *data.Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

// NewHeartbeatMessageInfo -
func NewHeartbeatMessageInfo(
	maxDurationPeerUnresponsive time.Duration,
	peerType string,
	genesisTime time.Time,
	timer heartbeat.Timer,
) (*heartbeatMessageInfo, error) {
	return newHeartbeatMessageInfo(
		maxDurationPeerUnresponsive,
		peerType,
		genesisTime,
		timer,
	)
}

// GetTimeStamp -
func (hbmi *heartbeatMessageInfo) GetTimeStamp() time.Time {
	return hbmi.timeStamp
}

// GetReceiverShardId -
func (hbmi *heartbeatMessageInfo) GetReceiverShardId() uint32 {
	return hbmi.receivedShardID
}

// GetTotalUpTime -
func (hbmi *heartbeatMessageInfo) GetTotalUpTime() time.Duration {
	return hbmi.totalUpTime
}

// GetComputedShardId -
func (hbmi *heartbeatMessageInfo) GetComputedShardId() uint32 {
	return hbmi.computedShardID
}

// GetPeerType -
func (hbmi *heartbeatMessageInfo) GetPeerType() string {
	return hbmi.peerType
}

// GetTotalDownTime -
func (hbmi *heartbeatMessageInfo) GetTotalDownTime() time.Duration {
	return hbmi.totalDownTime
}

// VerifyLengths -
func VerifyLengths(hbmi *data.Heartbeat) error {
	return verifyLengths(hbmi)
}

// GetMaxSizeInBytes -
func GetMaxSizeInBytes() int {
	return maxSizeInBytes
}

// GetNonce -
func (hbmi *heartbeatMessageInfo) GetNonce() uint64 {
	return hbmi.nonce
}

// RefreshHeartbeatMessageInfo -
func (m *Monitor) RefreshHeartbeatMessageInfo() {
	m.refreshHeartbeatMessageInfo()
}
