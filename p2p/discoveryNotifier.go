package p2p

import (
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

// DiscoveryNotifier implements the Notifee interface and is used to handle a new peer found
type DiscoveryNotifier struct {
	mes Messenger
}

// NewDiscoveryNotifier creates a new instance of DiscoveryNotifier struct
func NewDiscoveryNotifier(m Messenger) *DiscoveryNotifier {
	return &DiscoveryNotifier{mes: m}
}

// HandlePeerFound updates the routing table with this new peer
func (n *DiscoveryNotifier) HandlePeerFound(pi pstore.PeerInfo) {
	peers := n.mes.Peers()

	found := false

	for i := 0; i < len(peers); i++ {
		if peers[i] == pi.ID {
			found = true
			break
		}
	}

	if found {
		return
	}

	for i := 0; i < len(pi.Addrs); i++ {
		n.mes.AddAddr(pi.ID, pi.Addrs[i], pstore.PermanentAddrTTL)
	}

	n.mes.RouteTable().Update(pi.ID)
}
