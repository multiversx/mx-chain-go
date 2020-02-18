package discovery

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

type hostWithConnectionManagement struct {
	sharder Sharder
	ConnectableHost
}

// NewHostWithConnectionManagement returns a host wrapper able to decide if connection initiated to a peer
// will actually be kept or not
func NewHostWithConnectionManagement(ch ConnectableHost, sharder Sharder) (*hostWithConnectionManagement, error) {
	if ch == nil {
		return nil, p2p.ErrNilHost
	}

	return &hostWithConnectionManagement{
		ConnectableHost: ch,
		sharder:         sharder,
	}, nil
}

// Connect tries to connect connect to pi if the Connections Per Second (hd.cps)
// allows it.
func (hwcm *hostWithConnectionManagement) Connect(ctx context.Context, pi peer.AddrInfo) error {
	err := hwcm.checkIfCanConnectToPeer(pi.ID)
	if err != nil {
		return err
	}

	return hwcm.ConnectableHost.Connect(ctx, pi)
}

func (hwcm *hostWithConnectionManagement) checkIfCanConnectToPeer(pid peer.ID) error {
	sharder := hwcm.sharder
	if check.IfNil(sharder) {
		//no sharder in place, let them connect as usual
		return nil
	}

	allPeers := hwcm.ConnectableHost.Network().Peers()
	if hwcm.sharder.Has(pid, allPeers) {
		return fmt.Errorf("%w, pid: %s", p2p.ErrPeerAlreadyConnected, pid.Pretty())
	}

	allPeers = append(allPeers, pid)
	evicted := hwcm.sharder.ComputeEvictList(allPeers)
	if hwcm.sharder.Has(pid, evicted) {
		return fmt.Errorf("%w, pid: %s", p2p.ErrUnwantedPeer, pid.Pretty())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hwcm *hostWithConnectionManagement) IsInterfaceNil() bool {
	return hwcm == nil || check.IfNil(hwcm.ConnectableHost)
}
