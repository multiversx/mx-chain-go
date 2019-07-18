package heartbeat

import "time"

// heartbeatMessageInfo retain the message info received from another node (identified by a public key)
type heartbeatMessageInfo struct {
	maxDurationPeerUnresponsive time.Duration
	getTimeHandler              func() time.Time
	timeStamp                   time.Time
	maxInactiveTime             Duration
	isActive                    bool
	shardID                     uint32
	alreadyAccessed             bool
}

// newHeartbeatMessageInfo returns a new instance of a PubkeyElement
func newHeartbeatMessageInfo(maxDurationPeerUnresponsive time.Duration) (*heartbeatMessageInfo, error) {
	if maxDurationPeerUnresponsive == 0 {
		return nil, ErrInvalidMaxDurationPeerUnresponsive
	}

	hbmi := &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
	}

	hbmi.getTimeHandler = hbmi.clockTime
	hbmi.alreadyAccessed = false

	return hbmi, nil
}

func (hbmi *heartbeatMessageInfo) clockTime() time.Time {
	return time.Now()
}

// Sweep updates all records
func (hbmi *heartbeatMessageInfo) sweep() {
	crtDuration := hbmi.getTimeHandler().Sub(hbmi.timeStamp)
	hbmi.isActive = crtDuration < hbmi.maxDurationPeerUnresponsive
	if hbmi.maxInactiveTime.Duration < crtDuration {
		hbmi.maxInactiveTime.Duration = crtDuration
	}
}

// HeartbeatReceived processes a new message arrived from a p2p address
func (hbmi *heartbeatMessageInfo) HeartbeatReceived(shardID uint32) {
	crtTime := hbmi.getTimeHandler()
	hbmi.sweep()

	if !hbmi.alreadyAccessed {
		hbmi.shardID = shardID
		hbmi.isActive = true
		hbmi.timeStamp = crtTime
		hbmi.maxInactiveTime = Duration{Duration: 0}
	}
	crtDuration := crtTime.Sub(hbmi.timeStamp)
	if hbmi.maxInactiveTime.Duration < crtDuration {
		hbmi.maxInactiveTime.Duration = crtDuration
	}
	hbmi.timeStamp = crtTime
	hbmi.alreadyAccessed = true
}
