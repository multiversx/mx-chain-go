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

func (m *Monitor) GetExportedHbmi(hbmi *heartbeatMessageInfo) HeartbeatDTO {
	return m.convertToExportedStruct(hbmi)
}

func (m *Monitor) SendHeartbeatMessage(hb *Heartbeat) {
	m.sendHeartbeatMessage(hb)
}

func (m *Monitor) RestartMonitor() *Monitor {
	mon, _ := NewMonitor(
		m.singleSigner,
		m.keygen,
		m.marshalizer,
		m.maxDurationPeerUnresponsive,
		m.pubKeysMap,
		m.monitorDB,
		m.genesisTime)
	return mon
}
