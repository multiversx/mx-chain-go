package heartbeat

import "time"

// PubkeyElement retain the message info received from another node (identified by a public key)
type PubkeyElement struct {
	maxDurationPeerUnresponsive time.Duration
	m                           map[string]*PeerHeartbeat
	timeGetter                  func() time.Time
}

// NewPubkeyElement returns a new instance of a PubkeyElement
func NewPubkeyElement(maxDurationPeerUnresponsive time.Duration) *PubkeyElement {
	pe := &PubkeyElement{
		m:                           make(map[string]*PeerHeartbeat),
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
	}
	pe.timeGetter = pe.clockTimeGetter

	return pe
}

func (pe *PubkeyElement) clockTimeGetter() time.Time {
	return time.Now()
}

// Sweep updates all records
func (pe *PubkeyElement) sweep() {
	for _, phb := range pe.m {
		crtDuration := pe.timeGetter().Sub(phb.TimeStamp)
		phb.IsActive = crtDuration < pe.maxDurationPeerUnresponsive
		if phb.MaxInactiveTime.Duration < crtDuration {
			phb.MaxInactiveTime.Duration = crtDuration
		}
	}
}

// HeartbeatArrived processes a new message arrived from a p2p address
func (pe *PubkeyElement) HeartbeatArrived(p2pAddress string) {
	crtTime := pe.timeGetter()
	pe.sweep()

	phb := pe.m[p2pAddress]
	if phb == nil {
		pe.m[p2pAddress] = &PeerHeartbeat{
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
func (pe *PubkeyElement) GetPeerHeartbeats() []PeerHeartbeat {
	pe.sweep()
	heartbeats := make([]PeerHeartbeat, len(pe.m))

	idx := 0
	for _, phb := range pe.m {
		heartbeats[idx] = *phb
		idx++
	}

	return heartbeats
}
