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

	getTimeHandler         func() time.Time
	timeStamp              time.Time
	isActive               bool
	shardID                uint32
	versionNumber          string
	nodeDisplayName        string
	isValidator            bool
	lastTimeUptimeDowntime time.Time
}

// newHeartbeatMessageInfo returns a new instance of a PubkeyElement
func newHeartbeatMessageInfo(
	maxDurationPeerUnresponsive time.Duration,
	isValidator bool,
) (*heartbeatMessageInfo, error) {

	if maxDurationPeerUnresponsive == 0 {
		return nil, ErrInvalidMaxDurationPeerUnresponsive
	}

	hbmi := &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		maxInactiveTime:             Duration{0},
		isActive:                    false,
		timeStamp:                   emptyTimestamp,
		lastTimeUptimeDowntime:      time.Now(),
		totalUpTime:                 Duration{0},
		totalDownTime:               Duration{0},
		versionNumber:               "",
		nodeDisplayName:             "",
		isValidator:                 isValidator,
	}
	hbmi.getTimeHandler = hbmi.clockTime

	return hbmi, nil
}

func (hbmi *heartbeatMessageInfo) clockTime() time.Time {
	return time.Now()
}

func (hbmi *heartbeatMessageInfo) updateFields() {
	crtDuration := hbmi.getTimeHandler().Sub(hbmi.timeStamp)
	crtDuration = protectAgainstNegativeDuration(crtDuration)

	hbmi.isActive = crtDuration < hbmi.maxDurationPeerUnresponsive
	hbmi.updateUpAndDownTime()
	hbmi.updateMaxInactiveTimeDuration()
}

// Wil update the total time a node was up and down
func (hbmi *heartbeatMessageInfo) updateUpAndDownTime() {
	lastDuration := hbmi.clockTime().Sub(hbmi.lastTimeUptimeDowntime)
	lastDuration = protectAgainstNegativeDuration(lastDuration)

	if hbmi.isActive {
		hbmi.totalUpTime.Duration += lastDuration
	} else {
		hbmi.totalDownTime.Duration += lastDuration
	}

	hbmi.lastTimeUptimeDowntime = time.Now()
}

// HeartbeatReceived processes a new message arrived from a peer
func (hbmi *heartbeatMessageInfo) HeartbeatReceived(shardID uint32, version string, nodeDisplayName string) {
	crtTime := hbmi.getTimeHandler()
	hbmi.updateFields()
	hbmi.shardID = shardID
	hbmi.updateMaxInactiveTimeDuration()
	hbmi.timeStamp = crtTime
	hbmi.versionNumber = version
	hbmi.nodeDisplayName = nodeDisplayName
}

func (hbmi *heartbeatMessageInfo) updateMaxInactiveTimeDuration() {
	crtDuration := hbmi.getTimeHandler().Sub(hbmi.timeStamp)
	crtDuration = protectAgainstNegativeDuration(crtDuration)

	if hbmi.maxInactiveTime.Duration < crtDuration && hbmi.timeStamp != emptyTimestamp {
		hbmi.maxInactiveTime.Duration = crtDuration
	}
}

func protectAgainstNegativeDuration(src time.Duration) time.Duration {
	if src < 0 {
		return 0
	}

	return src
}
