package p2p

import (
	"bufio"
	"context"
	"crypto"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-metrics"
	libP2PNet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-swarm"
	tu "github.com/libp2p/go-libp2p-swarm/testing"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-tcp-transport"
	"github.com/pkg/errors"
)

type NetMessenger struct {
	p2pNode host.Host

	mutChansSend sync.RWMutex
	chansSend    map[string]chan []byte

	mutBootstrap sync.Mutex

	queue *MessageQueue
	marsh marshal.Marshalizer

	rt *RoutingTable
	cn *ConnNotifier

	mdns discovery.Service
	dn   *DiscoveryNotifier

	onMsgRecv func(caller Messenger, peerID string, m *Message)
	//OnMsgRecv          func(caller *Node, peerID string, m *Message)
}

func genSwarm(ctx context.Context, port int) (*swarm.Swarm, error) {
	p := randPeerNetParamsOrFatal(port)

	ps := pstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())
	ps.AddPubKey(p.ID, p.PubKey)
	ps.AddPrivKey(p.ID, p.PrivKey)
	s := swarm.NewSwarm(ctx, p.ID, ps, metrics.NewBandwidthCounter())

	tcpTransport := tcp.NewTCPTransport(tu.GenUpgrader(s))
	tcpTransport.DisableReuseport = false

	err := s.AddTransport(tcpTransport)
	if err != nil {
		return nil, err
	}

	err = s.Listen(p.Addr)
	if err != nil {
		return nil, err
	}

	s.Peerstore().AddAddrs(p.ID, s.ListenAddresses(), pstore.PermanentAddrTTL)

	return s, nil
}

func NewNetMessenger(ctx context.Context, mrsh marshal.Marshalizer, port int, addresses []string, maxAllowedPeers int) (*NetMessenger, error) {
	if mrsh == nil {
		return nil, errors.New("Marshalizer is nil! Can't create node!")
	}

	var node NetMessenger
	node.marsh = mrsh
	node.cn = NewConnNotifier(node)
	node.cn.MaxPeersAllowed = maxAllowedPeers

	timeStart := time.Now()

	netw, err := genSwarm(ctx, port)
	if err != nil {
		return nil, err
	}
	h := basichost.New(netw)

	fmt.Printf("Node: %v has the following addr table: \n", h.ID().Pretty())

	for i, addr := range h.Addrs() {
		fmt.Printf("%d: %s/ipfs/%s\n", i, addr, h.ID().Pretty())
	}
	fmt.Println()

	fmt.Printf("Created node in %v\n", time.Now().Sub(timeStart))

	node.p2pNode = h
	node.chansSend = make(map[string]chan []byte)
	node.queue = NewMessageQueue(50000)

	node.addRWHandlers("benchmark/nolimit/1.0.0.0")

	node.ConnectToAddresses(ctx, addresses)

	node.rt = NewRoutingTable(h.ID())
	//register the notifier
	node.p2pNode.Network().Notify(node.cn)
	node.cn.OnDoSimpleTask = func(this interface{}) {
		TaskMonitorConnections(node.cn)
	}
	node.cn.OnGetKnownPeers = func(cn *ConnNotifier) []peer.ID {
		return cn.Node.rt.NearestPeersAll()
	}
	node.cn.OnNeedToConnectToOtherPeer = func(cn *ConnNotifier, pid peer.ID) error {
		pinfo := node.P2pNode.Peerstore().PeerInfo(pid)

		if err := node.P2pNode.Connect(ctx, pinfo); err != nil {
			return err
		} else {
			stream, err := node.P2pNode.NewStream(ctx, pinfo.ID, "benchmark/nolimit/1.0.0.0")
			if err != nil {
				return err
			} else {
				node.StreamHandler(stream)
			}
		}

		return nil
	}

	return &node, nil
}

func (node *Node) ParseAddressIpfs(address string) (*pstore.PeerInfo, error) {
	addr, err := ipfsaddr.ParseString(address)
	if err != nil {
		return nil, err
	}

	pinfo, err := pstore.InfoFromP2pAddr(addr.Multiaddr())
	if err != nil {
		return nil, err
	}

	return pinfo, nil
}

func (node *NetMessenger) ConnectToAddresses(ctx context.Context, addresses []string) {
	peers := 0

	timeStart := time.Now()

	for i := 0; i < len(addresses); i++ {
		pinfo, err := node.ParseAddressIpfs(addresses[i])

		if err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		if err := node.P2pNode.Connect(ctx, *pinfo); err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		stream, err := node.P2pNode.NewStream(ctx, pinfo.ID, "benchmark/nolimit/1.0.0.0")
		if err != nil {
			fmt.Printf("Streaming the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		node.StreamHandler(stream)
		peers++
	}

	fmt.Printf("Connected to %d peers in %v\n", peers, time.Now().Sub(timeStart))
}

func randPeerNetParamsOrFatal(port int) ConnectParams {
	return *NewConnectParams(port)
}

func (node *NetMessenger) sendDirectRAW(peerID string, buff []byte) error {

	node.mutChansSend.RLock()
	chanSend, ok := node.chansSend[peerID]
	node.mutChansSend.RUnlock()

	if !ok {
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: fmt.Sprintf("Can not send to %v. Not connected?\n", peerID)}
	}

	select {
	case chanSend <- buff:
	default:
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: fmt.Sprintf("Can not send to %v. Pipe full! Message discarded!\n", peerID)}
	}

	return nil
}

func (node *NetMessenger) SendDirectBuff(peerID string, buff []byte) error {
	m := NewMessage(node.P2pNode.ID().Pretty(), buff, node.marsh)

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.sendDirectRAW(peerID, buff)
}

func (node *NetMessenger) SendDirectString(peerID string, message string) error {
	return node.SendDirectBuff(peerID, []byte(message))
}

func (node *NetMessenger) SendDirectMessage(peerID string, m *Message) error {
	if m == nil {
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: fmt.Sprintf("Can not send NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.sendDirectRAW(peerID, buff)
}

func (node *NetMessenger) broadcastRAW(buff []byte, excs []string) error {
	var errFound = &NodeError{}

	peers := node.P2pNode.Peerstore().Peers()

	for i := 0; i < len(peers); i++ {
		peerID := peer.ID(peers[i]).Pretty()

		if peerID == node.P2pNode.ID().Pretty() {
			continue
		}

		found := false
		for j := 0; j < len(excs); j++ {
			if peerID == excs[j] {
				found = true
				break
			}
		}

		if found {
			continue
		}

		err := node.sendDirectRAW(peerID, buff)

		if err != nil {
			errNode, _ := err.(*NodeError)
			errFound.NestedErrors = append(errFound.NestedErrors, *errNode)
		}
	}

	if len(errFound.NestedErrors) == 0 {
		return nil
	}

	if len(errFound.NestedErrors) == 1 {
		return &errFound.NestedErrors[0]
	}

	errFound.Err = "Multiple errors found!"
	return errFound
}

func (node *NetMessenger) BroadcastBuff(buff []byte, excs []string) error {
	m := NewMessage(node.P2pNode.ID().Pretty(), buff, node.marsh)

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.broadcastRAW(buff, excs)
}

func (node *NetMessenger) BroadcastString(message string, excs []string) error {
	return node.BroadcastBuff([]byte(message), excs)
}

func (node *NetMessenger) BroadcastMessage(m *Message, excs []string) error {
	if m == nil {
		return &NodeError{PeerRecv: "", PeerSend: node.P2pNode.ID().Pretty(), Err: fmt.Sprintf("Can not broadcast NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.broadcastRAW(buff, excs)
}

func (node *NetMessenger) StreamHandler(stream libP2PNet.Stream) {
	peerID := stream.Conn().RemotePeer().Pretty()

	node.mutChansSend.Lock()
	chanSend, ok := node.chansSend[peerID]
	if !ok {
		chanSend = make(chan []byte, 10000)
		node.chansSend[peerID] = chanSend
	}
	node.mutChansSend.Unlock()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go func(rw *bufio.ReadWriter, queue *MessageQueue) {
		for {
			buff, _ := rw.ReadBytes(byte('\n'))

			if len(buff) == 0 {
				continue
			}

			if (len(buff) == 1) && (buff[0] == byte('\n')) {
				continue
			}

			if buff[len(buff)-1] == byte('\n') {
				buff = buff[0 : len(buff)-1]
			}

			m, err := CreateFromByteArray(node.marsh, buff)

			if err != nil {
				continue
			}

			sha3 := crypto.SHA3_256.New()
			b64 := base64.StdEncoding
			hash := b64.EncodeToString(sha3.Sum([]byte(m.Payload)))

			if queue.ContainsAndAdd(hash) {
				continue
			}

			if node.onMsgRecv != nil {
				node.onMsgRecv(node, stream.Conn().RemotePeer().Pretty(), m)
			}
		}

	}(rw, node.queue)

	go func(rw *bufio.ReadWriter, chanSend chan []byte) {
		for {
			data := <-chanSend

			buff := make([]byte, len(data))
			copy(buff, data)

			buff = append(buff, byte('\n'))

			rw.Write(buff)
			rw.Flush()
		}
	}(rw, chanSend)
}

func (node *NetMessenger) addRWHandlers(protocolID protocol.ID) {
	node.p2pNode.SetStreamHandler(protocolID, node.StreamHandler)
}

func (node *NetMessenger) Bootstrap(ctx context.Context) {
	node.mutBootstrap.Lock()
	defer node.mutBootstrap.Unlock()

	if node.mdns != nil {
		return
	}

	node.dn = NewDiscoveryNotifier(node)

	mdns, err := discovery.NewMdnsService(context.Background(), node.p2pNode, time.Second, "discovery")

	if err != nil {
		panic(err)
	}

	mdns.RegisterNotifee(node.dn)
	node.mdns = mdns

	wait := time.Second * 10
	fmt.Printf("\n**** Waiting %v to bootstrap...****\n\n", wait)

	node.cn.Start()

	time.Sleep(wait)
}

func (node *NetMessenger) PrintConnected() {
	conns := node.p2pNode.Network().Conns()

	fmt.Printf("Node %s is connected to: \n", node.p2pNode.ID().Pretty())

	for i := 0; i < len(conns); i++ {
		fmt.Printf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(node.p2pNode.ID(), conns[i].RemotePeer()))
	}
}
