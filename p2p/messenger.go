package p2p

import (
	"context"
	_ "fmt"
	"io"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Messenger is used to communicate with another libP2P nodes
type Messenger interface {
	io.Closer

	ID() peer.ID
	Peers() []peer.ID
	Conns() []net.Conn
	Marshalizer() marshal.Marshalizer
	RouteTable() *RoutingTable
	Addrs() []string

	ConnectToAddresses(ctx context.Context, addresses []string)

	TopicHolder() *TopicHolder

	Bootstrap(ctx context.Context)
	PrintConnected()

	AddAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration)

	Connectedness(pid peer.ID) net.Connectedness
}
