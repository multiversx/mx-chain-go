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

type Messenger interface {
	io.Closer

	ID() peer.ID
	Peers() []peer.ID
	Conns() []net.Conn
	Marshalizer() marshal.Marshalizer
	RouteTable() *RoutingTable
	Addrs() []string

	OnRecvMsg() func(caller Messenger, peerID string, m *Message)
	SetOnRecvMsg(f func(caller Messenger, peerID string, m *Message))

	ConnectToAddresses(ctx context.Context, addresses []string)

	SendDirectBuff(peerID string, buff []byte) error
	SendDirectString(peerID string, message string) error
	SendDirectMessage(peerID string, m *Message) error

	BroadcastBuff(buff []byte, excs []string) error
	BroadcastString(message string, excs []string) error
	BroadcastMessage(m *Message, excs []string) error

	StreamHandler(stream net.Stream)
	Bootstrap(ctx context.Context)
	PrintConnected()

	AddAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration)

	Connectedness(pid peer.ID) net.Connectedness
}
