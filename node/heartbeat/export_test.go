package heartbeat

import "time"

func (m *Monitor) GetMessages() map[string]*heartbeatMessageInfo {
	return m.heartbeatMessages
}

func (m *Monitor) SetMessages(messages map[string]*heartbeatMessageInfo) {
	m.heartbeatMessages = messages
}

func (m *Monitor) GetHbmi(tmstp time.Time) *heartbeatMessageInfo {
	return &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: 0,
		maxInactiveTime:             Duration{},
		totalUpTime:                 Duration{},
		totalDownTime:               Duration{},
		getTimeHandler:              nil,
		timeStamp:                   time.Time{},
		isActive:                    false,
		receivedShardID:             0,
		computedShardID:             0,
		versionNumber:               "",
		nodeDisplayName:             "",
		isValidator:                 false,
		lastUptimeDowntime:          time.Time{},
		genesisTime:                 time.Time{},
	}
}

func (m *Monitor) SendHeartbeatMessage(hb *Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

func (m *Monitor) AddHeartbeatMessageToMap(hb *Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

func NewHeartbeatMessageInfo(
	maxDurationPeerUnresponsive time.Duration,
	isValidator bool,
	genesisTime time.Time,
	timer Timer,
) (*heartbeatMessageInfo, error) {
	return newHeartbeatMessageInfo(
		maxDurationPeerUnresponsive,
		isValidator,
		genesisTime,
		timer,
	)
}

func GetTimeStamp(hbmi *heartbeatMessageInfo) time.Time {
	return hbmi.timeStamp
}

func GetReceiverShardId(hbmi *heartbeatMessageInfo) uint32 {
	return hbmi.receivedShardID
}

func GetUpTimeDuration(hbmi *heartbeatMessageInfo) time.Duration {
	return hbmi.totalUpTime.Duration
}

func GetDownTimeDuration(hbmi *heartbeatMessageInfo) time.Duration {
	return hbmi.totalDownTime.Duration
}

func GetIsActive(hbmi *heartbeatMessageInfo) bool {
	return hbmi.isActive
}
