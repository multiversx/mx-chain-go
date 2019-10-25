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

func (hbmi *heartbeatMessageInfo) GetTimeStamp() time.Time {
	return hbmi.timeStamp
}

func (hbmi *heartbeatMessageInfo) GetReceiverShardId() uint32 {
	return hbmi.receivedShardID
}

func (hbmi *heartbeatMessageInfo) GetTotalUpTime() Duration {
	return hbmi.totalUpTime
}

func (hbmi *heartbeatMessageInfo) GetTotalDownTime() Duration {
	return hbmi.totalDownTime
}

func (hbmi *heartbeatMessageInfo) GetIsActive() bool {
	return hbmi.isActive
}
