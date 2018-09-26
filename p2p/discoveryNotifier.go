package p2p

import (
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

type DiscoveryNotifier struct {
	mes Messenger
}

func NewDiscoveryNotifier(m Messenger) *DiscoveryNotifier {
	return &DiscoveryNotifier{mes: m}
}

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
