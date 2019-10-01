package heartbeat

import (
	"time"
)

// heartbeatMessageInfo retain the message info received from another node (identified by a public key)
type heartbeatMessageInfo struct {
	maxDurationPeerUnresponsive time.Duration
	maxInactiveTime             Duration
	totalUpTime                 Duration
	totalDownTime               Duration

	getTimeHandler     func() time.Time
	timeStamp          time.Time
	isActive           bool
	receivedShardID    uint32
	computedShardID    uint32
	versionNumber      string
	nodeDisplayName    string
	isValidator        bool
	lastUptimeDowntime time.Time
	genesisTime        time.Time
}

// newHeartbeatMessageInfo returns a new instance of a heartbeatMessageInfo
func newHeartbeatMessageInfo(
	maxDurationPeerUnresponsive time.Duration,
	isValidator bool,
	genesisTime time.Time,
	getTimeHandler func() time.Time,
) (*heartbeatMessageInfo, error) {

	if maxDurationPeerUnresponsive == 0 {
		return nil, ErrInvalidMaxDurationPeerUnresponsive
	}
	if getTimeHandler == nil {
		return nil, ErrNilGetTimeHandler
	}

	hbmi := &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		maxInactiveTime:             Duration{0},
		isActive:                    false,
		receivedShardID:             uint32(0),
		timeStamp:                   genesisTime,
		lastUptimeDowntime:          getTimeHandler(),
		totalUpTime:                 Duration{0},
		totalDownTime:               Duration{0},
		versionNumber:               "",
		nodeDisplayName:             "",
		isValidator:                 isValidator,
		genesisTime:                 genesisTime,
		getTimeHandler:              getTimeHandler,
	}

	return hbmi, nil
}

func (hbmi *heartbeatMessageInfo) updateFields(crtTime time.Time) {
	if hbmi.genesisTime != hbmi.timeStamp {
		crtDuration := crtTime.Sub(hbmi.timeStamp)
		crtDuration = maxDuration(0, crtDuration)
		hbmi.isActive = crtDuration < hbmi.maxDurationPeerUnresponsive
		hbmi.updateMaxInactiveTimeDuration(crtTime)
		if crtTime.Sub(hbmi.genesisTime) > 0 {
			hbmi.updateUpAndDownTime(crtTime)
		}
	}
	hbmi.lastUptimeDowntime = crtTime
}

// Wil update the total time a node was up and down
func (hbmi *heartbeatMessageInfo) updateUpAndDownTime(crtTime time.Time) {
	lastDuration := crtTime.Sub(hbmi.lastUptimeDowntime)
	lastDuration = maxDuration(0, lastDuration)

	if hbmi.isActive {
		hbmi.totalUpTime.Duration += lastDuration
	} else {
		hbmi.totalDownTime.Duration += lastDuration
	}
}

// HeartbeatReceived processes a new message arrived from a peer
func (hbmi *heartbeatMessageInfo) HeartbeatReceived(computedShardID, receivedshardID uint32, version string,
	nodeDisplayName string) {
	crtTime := hbmi.getTimeHandler()
	hbmi.updateFields(crtTime)
	hbmi.computedShardID = computedShardID
	hbmi.receivedShardID = receivedshardID
	hbmi.updateMaxInactiveTimeDuration(crtTime)
	hbmi.timeStamp = crtTime
	hbmi.versionNumber = version
	hbmi.nodeDisplayName = nodeDisplayName
}

func (hbmi *heartbeatMessageInfo) updateMaxInactiveTimeDuration(currentTime time.Time) {
	crtDuration := currentTime.Sub(hbmi.timeStamp)
	crtDuration = maxDuration(0, crtDuration)

	if hbmi.maxInactiveTime.Duration < crtDuration && hbmi.timeStamp != hbmi.genesisTime {
		hbmi.maxInactiveTime.Duration = crtDuration
	}
}

func maxDuration(first, second time.Duration) time.Duration {
	if first > second {
		return first
	}

	return second
}
