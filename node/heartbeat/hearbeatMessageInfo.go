package heartbeat

import (
	"time"
)

var emptyTimestamp = time.Time{}

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
) (*heartbeatMessageInfo, error) {

	if maxDurationPeerUnresponsive == 0 {
		return nil, ErrInvalidMaxDurationPeerUnresponsive
	}

	hbmi := &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		maxInactiveTime:             Duration{0},
		isActive:                    false,
		receivedShardID:             uint32(0),
		timeStamp:                   genesisTime,
		lastUptimeDowntime:          time.Now(),
		totalUpTime:                 Duration{0},
		totalDownTime:               Duration{0},
		versionNumber:               "",
		nodeDisplayName:             "",
		isValidator:                 isValidator,
		genesisTime:                 genesisTime,
	}
	hbmi.getTimeHandler = hbmi.clockTime

	return hbmi, nil
}

func (hbmi *heartbeatMessageInfo) clockTime() time.Time {
	return time.Now()
}

func (hbmi *heartbeatMessageInfo) updateFields() {
	if hbmi.genesisTime != hbmi.timeStamp {
		crtDuration := hbmi.getTimeHandler().Sub(hbmi.timeStamp)
		crtDuration = maxDuration(0, crtDuration)
		hbmi.isActive = crtDuration < hbmi.maxDurationPeerUnresponsive
		hbmi.updateMaxInactiveTimeDuration()
		if time.Now().Sub(hbmi.genesisTime) > 0 {
			hbmi.updateUpAndDownTime()
		}
	}
	hbmi.lastUptimeDowntime = time.Now()
}

// Wil update the total time a node was up and down
func (hbmi *heartbeatMessageInfo) updateUpAndDownTime() {
	lastDuration := hbmi.clockTime().Sub(hbmi.lastUptimeDowntime)
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
	hbmi.updateFields()
	hbmi.computedShardID = computedShardID
	hbmi.receivedShardID = receivedshardID
	hbmi.updateMaxInactiveTimeDuration()
	hbmi.timeStamp = crtTime
	hbmi.versionNumber = version
	hbmi.nodeDisplayName = nodeDisplayName
}

func (hbmi *heartbeatMessageInfo) updateMaxInactiveTimeDuration() {
	crtDuration := hbmi.getTimeHandler().Sub(hbmi.timeStamp)
	crtDuration = maxDuration(0, crtDuration)

	if hbmi.maxInactiveTime.Duration < crtDuration && hbmi.timeStamp != emptyTimestamp {
		hbmi.maxInactiveTime.Duration = crtDuration
	}
}

func maxDuration(first, second time.Duration) time.Duration {
	if first > second {
		return first
	}

	return second
}
