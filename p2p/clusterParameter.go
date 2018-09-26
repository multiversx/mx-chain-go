package p2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

type ClusterParameter struct {
	ipAddress string
	startPort int
	endPort   int
	peers     []peer.ID
	addrs     []multiaddr.Multiaddr
}

func NewClusterParameter(ipAddr string, stPort int, enPort int) *ClusterParameter {
	cp := ClusterParameter{ipAddress: ipAddr, startPort: stPort, endPort: enPort}

	for i := stPort; i <= enPort; i++ {
		param := NewConnectParams(i)

		cp.peers = append(cp.peers, param.ID)
		ma, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d/ipfs/%s", ipAddr, i, param.ID.Pretty()))

		if err != nil {
			panic(err)
		}

		cp.addrs = append(cp.addrs, ma)
	}

	return &cp
}

func (cp *ClusterParameter) Peers() []peer.ID {
	return cp.peers
}

func (cp *ClusterParameter) Addrs() []multiaddr.Multiaddr {
	return cp.addrs
}
