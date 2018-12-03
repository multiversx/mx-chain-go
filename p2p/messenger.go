package p2p

import (
	"context"
	"io"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

var log = logger.NewDefaultLogger()

// Messenger is the main struct used for communicating with other peers
type Messenger interface {
	io.Closer

	ID() peer.ID
	Peers() []peer.ID
	Conns() []net.Conn
	Marshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	RouteTable() *RoutingTable
	Addresses() []string
	ConnectToAddresses(ctx context.Context, addresses []string)
	Bootstrap(ctx context.Context)
	PrintConnected()
	AddAddress(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration)
	Connectedness(pid peer.ID) net.Connectedness
	GetTopic(topicName string) *Topic
	AddTopic(t *Topic) error
}
