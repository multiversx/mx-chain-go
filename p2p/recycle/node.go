package recycle

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-floodsub"
	bhost "github.com/libp2p/go-libp2p-blankhost"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-swarm"
	tst "github.com/libp2p/go-libp2p-swarm/testing"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	"io"
)

type IncomingCallback func(*Node) net.StreamHandler

// Node .
type Node struct {
	Address    multiaddr.Multiaddr
	ID         peer.ID
	PS         peerstore.Peerstore
	Port       *int
	OutgoingID int
	host       host.Host
	PubSub     *floodsub.PubSub
	Context    context.Context
	PID        protocol.ID

	OnMsgRecvBroadcast func(sender *Node, topic string, msg *floodsub.Message)
}

func panicGuard(err error) {
	if err != nil {
		panic(err)
	}
}

// NewNode ..
func NewNode(r io.Reader, sourcePort *int, cb func(*Node) net.StreamHandler) (node *Node) {
	node = new(Node)
	node.Context = context.Background()
	node.PID = "/default"

	prvKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

	panicGuard(err)

	node.ID, _ = peer.IDFromPublicKey(pubKey)
	if sourcePort != nil {
		node.Port = sourcePort
		node.Address, _ = multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *node.Port))
	}

	node.PS = peerstore.NewPeerstore()
	node.PS.AddPrivKey(node.ID, prvKey)
	node.PS.AddPubKey(node.ID, pubKey)
	node.host = node.createHost(cb)

	node.PubSub, err = floodsub.NewFloodSub(node.Context, node.host)
	if err != nil {
		panic(err)
	}

	return
}

func (node *Node) createHost(cb IncomingCallback) host.Host {

	swarm := swarm.NewSwarm(node.Context, node.ID, node.PS, nil)

	tcpTransport := tcp.NewTCPTransport(tst.GenUpgrader(swarm))
	tcpTransport.DisableReuseport = false

	if err := swarm.AddTransport(tcpTransport); err != nil {
		panic(err)
	}

	h := bhost.NewBlankHost(swarm)
	if cb != nil {
		h.SetStreamHandler(node.PID, cb(node))
	}

	fmt.Println("Incomming address:")
	fmt.Printf("/ip4/127.0.0.1/tcp/%d/ipfs/%s\n", *node.Port, h.ID().Pretty())
	return h
}

// ConnectToDest .
func (node *Node) ConnectToDest(dest *string) *bufio.ReadWriter {
	if *dest == "" {
		return nil
	}

	peerID := addAddrToPeerstore(node.host, *dest)

	fmt.Println("This node's multiaddress: ")
	fmt.Printf("%s/ipfs/%s\n", node.Address, node.host.ID().Pretty())

	s, err := node.host.NewStream(node.Context, peerID, node.PID)

	panicGuard(err)

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	node.PS.Put(s.Conn().RemotePeer(), "stream", rw)

	return rw
}

func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	ipfsaddr, err := multiaddr.NewMultiaddr(addr)
	panicGuard(err)

	pid, err := ipfsaddr.ValueForProtocol(multiaddr.P_IPFS)
	panicGuard(err)

	peerid, err := peer.IDB58Decode(pid)
	panicGuard(err)

	targetPeerAddr, _ := multiaddr.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

func (node *Node) ConnectToAddress(addresses []string) {
	for _, address := range addresses {
		addr, err := ipfsaddr.ParseString(address)

		if err != nil {
			panic(err)
		}

		pinfo, _ := peerstore.InfoFromP2pAddr(addr.Multiaddr())

		if err := node.host.Connect(node.Context, *pinfo); err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed\n [%v]\n", address, err)
		}
	}
}

func (node *Node) ListenStreams(ctx context.Context, topic string) {
	sub, err := node.PubSub.Subscribe(topic)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				panic(err)
			}

			if node.OnMsgRecvBroadcast != nil {
				node.OnMsgRecvBroadcast(node, topic, msg)
			}
		}
	}()
}

// --- GO routines

//SendToPeers .
//func (node *Node) SendToPeers(message *Message) {
//	bytes, _ := json.Marshal(*message)
//	node.MS[message.Key()] = message
//	peers := node.PS.Peers()
//	for _, peer := range peers {
//		if peer != node.ID {
//			remoteRW, _ := node.PS.Get(peer, Stream)
//			rw := remoteRW.(*bufio.ReadWriter)
//			_, err := rw.Write(bytes)
//			rw.Flush()
//			panicGuard(err)
//		}
//	}

//type Node struct{
//	p2pNode host.Host
//	pubsub  *floodsub.PubSub
//}
//
//type config struct {
//	disableReuseport bool
//	dialOnly         bool
//}
//
//type Option func(*config)
//
//var OptDisableReuseport Option = func(c *config) {
//	c.disableReuseport = true
//}
//
//// OptDialOnly prevents the test swarm from listening.
//var OptDialOnly Option = func(c *config) {
//	c.dialOnly = true
//}
//
//func RandPeerNetParamsOrFatal(r io.Reader, port int) tu.PeerNetParams {
//	p, err := randPeerNetParams(r, port)
//	if err != nil {
//		panic(err)
//		return tu.PeerNetParams{} // TODO return nil
//	}
//	return *p
//}
//
//func randPeerNetParams(r io.Reader, port int) (*tu.PeerNetParams, error) {
//	var p tu.PeerNetParams
//	var err error
//	p.Addr,_ = ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
//	p.PrivKey, p.PubKey, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
//	if err != nil {
//		return nil, err
//	}
//	p.ID, err = peer.IDFromPublicKey(p.PubKey)
//	if err != nil {
//		return nil, err
//	}
//	return &p, nil
//}
//
//func GenSwarm(r io.Reader, ctx context.Context, port int, opts ...Option) *swarm.Swarm {
//	var cfg config
//	for _, o := range opts {
//		o(&cfg)
//	}
//
//	p := RandPeerNetParamsOrFatal(r, port)
//
//	ps := pstore.NewPeerstore()
//	ps.AddPubKey(p.ID, p.PubKey)
//	ps.AddPrivKey(p.ID, p.PrivKey)
//
//	swarm := swarm.NewSwarm(context.Background(), p.ID, ps, nil)
//	swarm.Listen(p.Addr)
//	swarm.Peerstore().AddAddr(p.ID, p.Addr, pstore.PermanentAddrTTL)
//
//	return swarm
//}
//
//func CreateNewNode(r io.Reader, ctx context.Context, addrToConnect []string, port int) *Node {
//	var node Node
//
//	netw := GenSwarm(r, ctx, port)
//	h := bhost.NewBlankHost(netw)
//
//	pubsub, err := floodsub.NewFloodSub(ctx, h)
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Printf("Node: %v has the following addr table: \n", h.ID().Pretty())
//
//	for i, addr := range h.Addrs() {
//		fmt.Printf("%d: %s/ipfs/%s\n", i, addr, h.ID().Pretty())
//	}
//	fmt.Println()
//
//	for _, address := range(addrToConnect){
//		addr, err := ipfsaddr.ParseString(address)
//		if err != nil {
//			panic(err)
//		}
//		pinfo, _ := peerstore.InfoFromP2pAddr(addr.Multiaddr())
//
//		if err := h.Connect(ctx, *pinfo); err != nil {
//			fmt.Println("bootstrapping a peer failed", err)
//		}
//
//
//	}
//
//	node.p2pNode = h
//	node.pubsub = pubsub
//
//	node.ListenStreams(ctx, "data")
//
//	return &node
//}
//
//func (node *Node) ListenStreams(ctx context.Context, topic string) {
//	sub, err := node.pubsub.Subscribe(topic)
//	if err != nil {
//		panic(err)
//	}
//
//	go func() {
//		for {
//			msg, err := sub.Next(ctx)
//			if err != nil {
//				panic(err)
//			}
//
//			fmt.Printf("%v got on topic %v, message: %v\n", node.p2pNode.ID().Pretty(), topic, msg.GetData())
//		}
//	}()
//}
