package heartbeat

import "time"

// heartbeatMessageInfo retain the message info received from another node (identified by a public key)
type heartbeatMessageInfo struct {
	maxDurationPeerUnresponsive time.Duration
	peerHeartbeats              map[string]*PeerHeartbeat
	getTimeHandler              func() time.Time
}

// newHeartbeatMessageInfo returns a new instance of a PubkeyElement
func newHeartbeatMessageInfo(maxDurationPeerUnresponsive time.Duration) (*heartbeatMessageInfo, error) {
	if maxDurationPeerUnresponsive == 0 {
		return nil, ErrInvalidMaxDurationPeerUnresponsive
	}

	hbmi := &heartbeatMessageInfo{
		peerHeartbeats:              make(map[string]*PeerHeartbeat),
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
	}
	hbmi.getTimeHandler = hbmi.clockTime

	return hbmi, nil
}

func (hbmi *heartbeatMessageInfo) clockTime() time.Time {
	return time.Now()
}

// Sweep updates all records
func (hbmi *heartbeatMessageInfo) sweep() {
	for _, phb := range hbmi.peerHeartbeats {
		crtDuration := hbmi.getTimeHandler().Sub(phb.TimeStamp)
		phb.IsActive = crtDuration < hbmi.maxDurationPeerUnresponsive
		if phb.MaxInactiveTime.Duration < crtDuration {
			phb.MaxInactiveTime.Duration = crtDuration
		}
	}
}

// HeartbeatReceived processes a new message arrived from a p2p address
func (hbmi *heartbeatMessageInfo) HeartbeatReceived(p2pAddress string) {
	crtTime := hbmi.getTimeHandler()
	hbmi.sweep()

	phb := hbmi.peerHeartbeats[p2pAddress]
	if phb == nil {
		hbmi.peerHeartbeats[p2pAddress] = &PeerHeartbeat{
			P2PAddress:      p2pAddress,
			TimeStamp:       crtTime,
			MaxInactiveTime: Duration{Duration: 0},
			IsActive:        true,
		}
		return
	}

	phb.IsActive = true
	crtDuration := crtTime.Sub(phb.TimeStamp)
	if phb.MaxInactiveTime.Duration < crtDuration {
		phb.MaxInactiveTime.Duration = crtDuration
	}
	phb.TimeStamp = crtTime
}

// GetPeerHeartbeats returns the updated peer heart beats collection
func (hbmi *heartbeatMessageInfo) GetPeerHeartbeats() []PeerHeartbeat {
	hbmi.sweep()
	heartbeats := make([]PeerHeartbeat, len(hbmi.peerHeartbeats))

	idx := 0
	for _, phb := range hbmi.peerHeartbeats {
		heartbeats[idx] = *phb
		idx++
	}

	return heartbeats
}
