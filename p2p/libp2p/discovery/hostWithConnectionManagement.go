package discovery

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

type hostWithConnectionManagement struct {
	sharder libp2p.Sharder
	ConnectableHost
}

// NewHostWithConnectionManagement returns a host wrapper able to decide if connection initiated to a peer
// will actually be kept or not
func NewHostWithConnectionManagement(ch ConnectableHost, sharder libp2p.Sharder) (*hostWithConnectionManagement, error) {
	if check.IfNil(ch) {
		return nil, p2p.ErrNilHost
	}
	//TODO add check if nil for sharder after the refactoring is done

	return &hostWithConnectionManagement{
		ConnectableHost: ch,
		sharder:         sharder,
	}, nil
}

// Connect tries to connect to the provided address info if the sharder allows it
func (hwcm *hostWithConnectionManagement) Connect(ctx context.Context, pi peer.AddrInfo) error {
	err := hwcm.canConnectToPeer(pi.ID)
	if err != nil {
		return err
	}

	return hwcm.ConnectableHost.Connect(ctx, pi)
}

func (hwcm *hostWithConnectionManagement) canConnectToPeer(pid peer.ID) error {
	sharder := hwcm.sharder
	if check.IfNil(sharder) {
		//no sharder in place, let them connect as usual
		return nil
	}

	allPeers := hwcm.ConnectableHost.Network().Peers()
	if !hwcm.sharder.Has(pid, allPeers) {
		allPeers = append(allPeers, pid)
	}

	evicted := hwcm.sharder.ComputeEvictionList(allPeers)
	if hwcm.sharder.Has(pid, evicted) {
		return fmt.Errorf("%w, pid: %s", p2p.ErrUnwantedPeer, pid.Pretty())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hwcm *hostWithConnectionManagement) IsInterfaceNil() bool {
	return hwcm == nil || check.IfNil(hwcm.ConnectableHost)
}
