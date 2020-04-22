package heartbeat

import (
	"time"
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

// GetHbmi -
func (m *Monitor) GetHbmi(_ time.Time) *heartbeatMessageInfo {
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
func (m *Monitor) SendHeartbeatMessage(hb *Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

// AddHeartbeatMessageToMap -
func (m *Monitor) AddHeartbeatMessageToMap(hb *Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

// NewHeartbeatMessageInfo -
func NewHeartbeatMessageInfo(
	maxDurationPeerUnresponsive time.Duration,
	peerType string,
	genesisTime time.Time,
	timer Timer,
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

// GetTotalDownTime -
func (hbmi *heartbeatMessageInfo) GetTotalDownTime() time.Duration {
	return hbmi.totalDownTime
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
func VerifyLengths(hbmi *Heartbeat) error {
	return verifyLengths(hbmi)
}

// GetMaxSizeInBytes -
func GetMaxSizeInBytes() int {
	return maxSizeInBytes
}
