package p2p

import (
	"bufio"
	"context"
	"crypto"
	"encoding/base64"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-floodsub"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-metrics"
	"github.com/libp2p/go-libp2p-net"
	libP2PNet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-swarm"
	tu "github.com/libp2p/go-libp2p-swarm/testing"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-tcp-transport"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Node struct {
	P2pNode host.Host

	mutConnected sync.RWMutex
	connected    map[peer.ID]string

	mutChansSend sync.RWMutex
	chansSend    map[string]chan []byte

	mutBootstrap sync.Mutex

	queue *MessageQueue
	marsh marshal.Marshalizer

	rt *RoutingTable

	OnMsgRecvBroadcast func(sender *Node, topic string, msg *floodsub.Message)
	OnMsgRecv          func(sender *Node, peerID string, m *Message)

	MaxAllowedPeers int
}

func GenSwarm(ctx context.Context, port int) (*swarm.Swarm, error) {
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

func NewNode(ctx context.Context, port int, addresses []string, mrsh marshal.Marshalizer) (*Node, error) {
	if mrsh == nil {
		return nil, errors.New("Marshalizer is nil! Can't create node!")
	}

	var node Node
	node.marsh = mrsh
	node.MaxAllowedPeers = 10
	node.connected = make(map[peer.ID]string)

	timeStart := time.Now()

	netw, err := GenSwarm(ctx, port)
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

	node.P2pNode = h
	node.chansSend = make(map[string](chan []byte))
	node.queue = NewMessageQueue(50000)

	node.addRWHandlers("benchmark/nolimit/1.0.0.0")

	node.ConnectToAddresses(ctx, addresses)

	node.rt = NewRoutingTable(h.ID())

	return &node, nil
}

func (node *Node) ConnectToAddresses(ctx context.Context, addresses []string) {
	peers := 0

	timeStart := time.Now()

	for _, address := range addresses {

		node.mutConnected.RLock()
		var len = len(node.connected)
		node.mutConnected.RUnlock()

		if len >= node.MaxAllowedPeers {
			return
		}

		addr, err := ipfsaddr.ParseString(address)
		if err != nil {
			panic(err)
		}
		pinfo, _ := pstore.InfoFromP2pAddr(addr.Multiaddr())

		if err := node.P2pNode.Connect(ctx, *pinfo); err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", address, err)
		} else {
			stream, err := node.P2pNode.NewStream(ctx, pinfo.ID, "benchmark/nolimit/1.0.0.0")
			if err != nil {
				fmt.Printf("Streaming the peer '%v' failed with error %v\n", address, err)
			} else {
				node.streamHandler(stream)
				peers++
			}
		}
	}

	fmt.Printf("Connected to %d peers in %v\n", peers, time.Now().Sub(timeStart))
}

func randPeerNetParamsOrFatal(port int) P2PParams {
	return *NewP2PParams(port)
}

func (node *Node) sendDirectRAW(peerID string, buff []byte) error {

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

func (node *Node) SendDirectBuff(peerID string, buff []byte) error {
	m := NewMessage(node.P2pNode.ID().Pretty(), buff, node.marsh)

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.sendDirectRAW(peerID, buff)
}

func (node *Node) SendDirectString(peerID string, message string) error {
	return node.SendDirectBuff(peerID, []byte(message))
}

func (node *Node) SendDirectMessage(peerID string, m *Message) error {
	if m == nil {
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: fmt.Sprintf("Can not send NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.sendDirectRAW(peerID, buff)
}

func (node *Node) broadcastRAW(buff []byte, excs []string) error {
	var errFound = &NodeError{}

	for _, pid := range node.P2pNode.Peerstore().Peers() {
		peerID := peer.ID(pid).Pretty()

		if peerID == node.P2pNode.ID().Pretty() {
			continue
		}

		found := false
		for _, exc := range excs {
			if peerID == exc {
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

func (node *Node) BroadcastBuff(buff []byte, excs []string) error {
	m := NewMessage(node.P2pNode.ID().Pretty(), buff, node.marsh)

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.broadcastRAW(buff, excs)
}

func (node *Node) BroadcastString(message string, excs []string) error {
	return node.BroadcastBuff([]byte(message), excs)
}

func (node *Node) BroadcastMessage(m *Message, excs []string) error {
	if m == nil {
		return &NodeError{PeerRecv: "", PeerSend: node.P2pNode.ID().Pretty(), Err: fmt.Sprintf("Can not broadcast NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: node.P2pNode.ID().Pretty(), Err: err.Error()}
	}

	return node.broadcastRAW(buff, excs)
}

func (node *Node) streamHandler(stream libP2PNet.Stream) {
	peerID := stream.Conn().RemotePeer().Pretty()

	node.mutConnected.Lock()
	v, found := node.connected[stream.Conn().RemotePeer()]

	if stream.Stat().Direction == net.DirInbound {
		if !found {
			node.connected[stream.Conn().RemotePeer()] = "I"
		} else {
			if v == "O" {
				node.connected[stream.Conn().RemotePeer()] = "I/O"
			}
		}
	} else {
		if stream.Stat().Direction == net.DirOutbound {
			if !found {
				node.connected[stream.Conn().RemotePeer()] = "O"
			} else {
				if v == "I" {
					node.connected[stream.Conn().RemotePeer()] = "I/O"
				}
			}
		}
	}
	node.mutConnected.Unlock()

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
			base64 := base64.StdEncoding
			hash := base64.EncodeToString(sha3.Sum([]byte(m.Payload)))

			if queue.ContainsAndAdd(hash) {
				continue
			}

			if node.OnMsgRecv != nil {
				node.OnMsgRecv(node, stream.Conn().RemotePeer().Pretty(), m)
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

func (node *Node) addRWHandlers(protocolID protocol.ID) {
	node.P2pNode.SetStreamHandler(protocolID, node.streamHandler)
}

func (node *Node) Bootstrap(ctx context.Context, cps []ClusterParameter, nConns int) {
	node.mutBootstrap.Lock()
	defer node.mutBootstrap.Unlock()

	for _, cp := range cps {
		for idx, peer := range cp.Peers() {
			str := cp.Addrs()[idx].String()

			addr, err := ipfsaddr.ParseString(str)

			if err != nil {
				panic(err)
			}

			pinfo, err := pstore.InfoFromP2pAddr(addr.Multiaddr())

			if err != nil {
				panic(err)
			}

			ma := pinfo.Addrs[0]

			node.P2pNode.Peerstore().AddAddr(peer, ma, pstore.PermanentAddrTTL)
			node.rt.Update(peer)
		}
	}

	peersToConnect := node.rt.NearestPeers(nConns)

	for _, peer := range peersToConnect {
		pinfo := node.P2pNode.Peerstore().PeerInfo(peer)

		if err := node.P2pNode.Connect(ctx, pinfo); err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", pinfo.Addrs, err)
		} else {
			stream, err := node.P2pNode.NewStream(ctx, pinfo.ID, "benchmark/nolimit/1.0.0.0")
			if err != nil {
				fmt.Printf("Streaming the peer '%v' failed with error %v\n", pinfo.Addrs, err)
			} else {
				node.streamHandler(stream)
			}
		}
	}

}

func (node *Node) PrintConnected() {
	fmt.Printf("Node %s has the following connections:\n", node.P2pNode.ID().Pretty())
	node.mutConnected.Lock()
	defer node.mutConnected.Unlock()

	for k, v := range node.connected {
		fmt.Printf("\t - %v type %s, distance %d\n", k, v, ComputeDistanceAD(node.P2pNode.ID(), k))
	}
}
