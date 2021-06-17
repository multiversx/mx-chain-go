package discovery

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type hostWithConnectionManagement struct {
	sharder Sharder
	ConnectableHost
}

// NewHostWithConnectionManagement returns a host wrapper able to decide if connection initiated to a peer
// will actually be kept or not
func NewHostWithConnectionManagement(ch ConnectableHost, sharder Sharder) (*hostWithConnectionManagement, error) {
	if check.IfNil(ch) {
		return nil, p2p.ErrNilHost
	}
	if check.IfNil(sharder) {
		return nil, p2p.ErrNilSharder
	}

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

// IsConnected returns true if the current host is connected to the provided peer info
func (hwcm *hostWithConnectionManagement) IsConnected(pi peer.AddrInfo) bool {
	return hwcm.Network().Connectedness(pi.ID) == network.Connected
}

// IsInterfaceNil returns true if there is no value under the interface
func (hwcm *hostWithConnectionManagement) IsInterfaceNil() bool {
	return hwcm == nil || check.IfNil(hwcm.ConnectableHost)
}
