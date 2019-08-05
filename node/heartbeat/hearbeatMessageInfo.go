package heartbeat

import (
	"time"
)

var emptyTimestamp = time.Time{}

// heartbeatMessageInfo retain the message info received from another node (identified by a public key)
type heartbeatMessageInfo struct {
	maxDurationPeerUnresponsive time.Duration
	getTimeHandler              func() time.Time
	timeStamp                   time.Time
	maxInactiveTime             Duration
	isActive                    bool
	shardID                     uint32
	totalUpTime                 Duration
	totalDownTime               Duration
}

// newHeartbeatMessageInfo returns a new instance of a PubkeyElement
func newHeartbeatMessageInfo(maxDurationPeerUnresponsive time.Duration) (*heartbeatMessageInfo, error) {
	if maxDurationPeerUnresponsive == 0 {
		return nil, ErrInvalidMaxDurationPeerUnresponsive
	}

	hbmi := &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		maxInactiveTime:             Duration{0},
		isActive:                    false,
	}
	hbmi.getTimeHandler = hbmi.clockTime
	hbmi.timeStamp = emptyTimestamp
	hbmi.totalUpTime = Duration{0}
	hbmi.totalDownTime = Duration{0}

	return hbmi, nil
}

func (hbmi *heartbeatMessageInfo) clockTime() time.Time {
	return time.Now()
}

// Sweep updates all records
func (hbmi *heartbeatMessageInfo) sweep() {
	crtDuration := hbmi.getTimeHandler().Sub(hbmi.timeStamp)
	hbmi.isActive = crtDuration < hbmi.maxDurationPeerUnresponsive
	hbmi.updateUpAndDownTime()
	hbmi.updateMaxInactiveTimeDuration()
}

// Wil update the total time a node was up and down
func (hbmi *heartbeatMessageInfo) updateUpAndDownTime() {
	if hbmi.isActive {
		hbmi.totalUpTime.Duration += hbmi.clockTime().Sub(hbmi.timeStamp)
	} else {
		if hbmi.timeStamp != emptyTimestamp {
			hbmi.totalDownTime.Duration += hbmi.clockTime().Sub(hbmi.timeStamp)
		}
	}
}

// HeartbeatReceived processes a new message arrived from a peer
func (hbmi *heartbeatMessageInfo) HeartbeatReceived(shardID uint32) {
	crtTime := hbmi.getTimeHandler()
	hbmi.sweep()
	hbmi.shardID = shardID
	hbmi.updateMaxInactiveTimeDuration()
	hbmi.timeStamp = crtTime
}

func (hbmi *heartbeatMessageInfo) updateMaxInactiveTimeDuration() {
	crtDuration := hbmi.getTimeHandler().Sub(hbmi.timeStamp)
	if hbmi.maxInactiveTime.Duration < crtDuration && hbmi.timeStamp != emptyTimestamp {
		hbmi.maxInactiveTime.Duration = crtDuration
	}
}
