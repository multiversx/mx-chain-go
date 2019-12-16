package antiflood

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type p2pAntiflood struct {
	p2p.FloodPreventer
}

// NewP2pAntiflood creates a new p2p anti flood protection mechanism built on top of a flood preventer implementation.
// It contains only the p2p anti flood logic that should be applied
func NewP2pAntiflood(floodPreventer p2p.FloodPreventer) (*p2pAntiflood, error) {
	if check.IfNil(floodPreventer) {
		return nil, p2p.ErrNilFloodPreventer
	}

	return &p2pAntiflood{
		FloodPreventer: floodPreventer,
	}, nil
}

// CanProcessMessage signals if a p2p message can or not be processed
func (af *p2pAntiflood) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	floodPreventer := af.FloodPreventer
	if check.IfNil(floodPreventer) {
		return p2p.ErrNilFloodPreventer
	}
	if message == nil {
		return p2p.ErrNilMessage
	}

	//protect from directly connected peer
	ok := floodPreventer.Increment(fromConnectedPeer.Pretty(), uint64(len(message.Data())))
	if !ok {
		return fmt.Errorf("%w in p2pAntiflood for connected peer", p2p.ErrSystemBusy)
	}

	if fromConnectedPeer != message.Peer() {
		//protect from the flooding messages that originate from the same source but come from different peers
		ok = floodPreventer.Increment(message.Peer().Pretty(), uint64(len(message.Data())))
		if !ok {
			return fmt.Errorf("%w in p2pAntiflood for originator", p2p.ErrSystemBusy)
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *p2pAntiflood) IsInterfaceNil() bool {
	return af == nil || check.IfNil(af.FloodPreventer)
}
