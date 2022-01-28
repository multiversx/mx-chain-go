package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// ArgsHostWithConnectionManagement is the argument DTO used in the NewHostWithConnectionManagement function
type ArgsHostWithConnectionManagement struct {
	ConnectableHost    ConnectableHost
	Sharder            Sharder
	ConnectionsWatcher p2p.ConnectionsWatcher
}

type hostWithConnectionManagement struct {
	ConnectableHost
	sharder            Sharder
	connectionsWatcher p2p.ConnectionsWatcher
}

// NewHostWithConnectionManagement returns a host wrapper able to decide if connection initiated to a peer
// will actually be kept or not
func NewHostWithConnectionManagement(args ArgsHostWithConnectionManagement) (*hostWithConnectionManagement, error) {
	if check.IfNil(args.ConnectableHost) {
		return nil, p2p.ErrNilHost
	}
	if check.IfNil(args.Sharder) {
		return nil, p2p.ErrNilSharder
	}
	if check.IfNil(args.ConnectionsWatcher) {
		return nil, p2p.ErrNilConnectionsWatcher
	}

	return &hostWithConnectionManagement{
		ConnectableHost:    args.ConnectableHost,
		sharder:            args.Sharder,
		connectionsWatcher: args.ConnectionsWatcher,
	}, nil
}

// Connect tries to connect to the provided address info if the sharder allows it
func (hwcm *hostWithConnectionManagement) Connect(ctx context.Context, pi peer.AddrInfo) error {
	addresses := concatenateAddresses(pi.Addrs)
	hwcm.connectionsWatcher.NewKnownConnection(core.PeerID(pi.ID), addresses)
	err := hwcm.canConnectToPeer(pi.ID)
	if err != nil {
		return err
	}

	return hwcm.ConnectableHost.Connect(ctx, pi)
}

func concatenateAddresses(addresses []multiaddr.Multiaddr) string {
	sb := strings.Builder{}
	for _, ma := range addresses {
		sb.WriteString(ma.String() + " ")
	}

	return sb.String()
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
