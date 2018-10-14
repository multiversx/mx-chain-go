package p2p

import (
	"bufio"
	"context"
	"crypto"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-conn-security-multistream"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-metrics"
	libP2PNet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-secio"
	"github.com/libp2p/go-libp2p-swarm"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	msmux "github.com/whyrusleeping/go-smux-multistream"
	yamux "github.com/whyrusleeping/go-smux-yamux"
)

type NetMessenger struct {
	protocol protocol.ID

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

	mutClosed sync.RWMutex
	closed    bool
}

func genUpgrader(n *swarm.Swarm) *tptu.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)

	return &tptu.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: n.Filters,
	}

}

func genSwarm(ctx context.Context, cp ConnectParams) (*swarm.Swarm, error) {
	ps := pstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())
	ps.AddPubKey(cp.ID, cp.PubKey)
	ps.AddPrivKey(cp.ID, cp.PrivKey)
	s := swarm.NewSwarm(ctx, cp.ID, ps, metrics.NewBandwidthCounter())

	tcpTransport := tcp.NewTCPTransport(genUpgrader(s))
	tcpTransport.DisableReuseport = false

	err := s.AddTransport(tcpTransport)
	if err != nil {
		return nil, err
	}

	err = s.Listen(cp.Addr)
	if err != nil {
		return nil, err
	}

	s.Peerstore().AddAddrs(cp.ID, s.ListenAddresses(), pstore.PermanentAddrTTL)

	return s, nil
}

func NewNetMessenger(ctx context.Context, mrsh marshal.Marshalizer, cp ConnectParams, addresses []string, maxAllowedPeers int) (*NetMessenger, error) {
	if mrsh == nil {
		return nil, errors.New("marshalizer is nil! Can't create node")
	}

	var node NetMessenger
	node.marsh = mrsh
	node.cn = NewConnNotifier(&node)
	node.cn.MaxAllowedPeers = maxAllowedPeers

	timeStart := time.Now()

	netw, err := genSwarm(ctx, cp)
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
	node.protocol = protocol.ID("elrondnetwork/1.0.0.0")

	node.p2pNode.SetStreamHandler(node.protocol, node.StreamHandler)

	node.ConnectToAddresses(ctx, addresses)

	node.rt = NewRoutingTable(h.ID())
	//register the notifier
	node.p2pNode.Network().Notify(node.cn)
	node.cn.OnDoSimpleTask = func(this interface{}) {
		TaskMonitorConnections(node.cn)
	}
	node.cn.OnGetKnownPeers = func(cn *ConnNotifier) []peer.ID {
		return cn.Msgr.RouteTable().NearestPeersAll()
	}
	node.cn.OnNeedToConn = func(cn *ConnNotifier, pid peer.ID) error {
		pinfo := node.p2pNode.Peerstore().PeerInfo(pid)

		if err := node.p2pNode.Connect(ctx, pinfo); err != nil {
			return err
		} else {
			stream, err := node.p2pNode.NewStream(ctx, pinfo.ID, node.protocol)
			if err != nil {
				return err
			} else {
				node.StreamHandler(stream)
			}
		}

		return nil
	}

	node.mutClosed = sync.RWMutex{}
	node.closed = false

	return &node, nil
}

func (nm *NetMessenger) Close() error {
	nm.mutClosed.Lock()
	nm.closed = true
	nm.mutClosed.Unlock()

	nm.p2pNode.Close()

	return nil
}

func (nm *NetMessenger) ID() peer.ID {
	return nm.p2pNode.ID()
}

func (nm *NetMessenger) Peers() []peer.ID {
	return nm.p2pNode.Peerstore().Peers()
}

func (nm *NetMessenger) Conns() []libP2PNet.Conn {
	return nm.p2pNode.Network().Conns()
}

func (nm *NetMessenger) Marshalizer() marshal.Marshalizer {
	return nm.marsh
}

func (nm *NetMessenger) RouteTable() *RoutingTable {
	return nm.rt
}

func (nm *NetMessenger) Addrs() []string {
	addrs := make([]string, 0)

	for _, adrs := range nm.p2pNode.Addrs() {
		addrs = append(addrs, adrs.String()+"/ipfs/"+nm.ID().Pretty())
	}

	return addrs
}

func (nm *NetMessenger) OnRecvMsg() func(caller Messenger, peerID string, m *Message) {
	return nm.onMsgRecv
}

func (nm *NetMessenger) SetOnRecvMsg(f func(caller Messenger, peerID string, m *Message)) {
	nm.onMsgRecv = f
}

func (nm *NetMessenger) ConnectToAddresses(ctx context.Context, addresses []string) {
	peers := 0

	timeStart := time.Now()

	for i := 0; i < len(addresses); i++ {
		pinfo, err := nm.ParseAddressIpfs(addresses[i])

		if err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		if err := nm.p2pNode.Connect(ctx, *pinfo); err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		stream, err := nm.p2pNode.NewStream(ctx, pinfo.ID, nm.protocol)
		if err != nil {
			fmt.Printf("Streaming the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		nm.StreamHandler(stream)
		peers++
	}

	fmt.Printf("Connected to %d peers in %v\n", peers, time.Now().Sub(timeStart))
}

func (nm *NetMessenger) sendDirectRAW(peerID string, buff []byte) error {
	nm.mutClosed.RLock()
	if nm.closed {
		nm.mutClosed.RUnlock()
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: "Attempt to write on a closed messenger!\n"}
	}
	nm.mutClosed.RUnlock()

	if peerID == nm.ID().Pretty() {
		//send to self allowed
		nm.gotNewMessage(buff, nm.ID())

		return nil
	}

	nm.mutChansSend.RLock()
	chanSend, ok := nm.chansSend[peerID]
	nm.mutChansSend.RUnlock()

	if !ok {
		return &NodeError{PeerRecv: peerID, PeerSend: nm.ID().Pretty(), Err: fmt.Sprintf("Can not send to %v. Not connected?\n", peerID)}
	}

	select {
	case chanSend <- buff:
	default:
		return &NodeError{PeerRecv: peerID, PeerSend: nm.ID().Pretty(), Err: fmt.Sprintf("Can not send to %v. Pipe full! Message discarded!\n", peerID)}
	}

	return nil
}

func (nm *NetMessenger) SendDirectBuff(peerID string, buff []byte) error {
	m := NewMessage(nm.ID().Pretty(), buff, nm.Marshalizer())

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: nm.ID().Pretty(), Err: err.Error()}
	}

	return nm.sendDirectRAW(peerID, buff)
}

func (nm *NetMessenger) SendDirectString(peerID string, message string) error {
	return nm.SendDirectBuff(peerID, []byte(message))
}

func (nm *NetMessenger) SendDirectMessage(peerID string, m *Message) error {
	if m == nil {
		return &NodeError{PeerRecv: peerID, PeerSend: nm.ID().Pretty(), Err: fmt.Sprintf("Can not send NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: nm.ID().Pretty(), Err: err.Error()}
	}

	return nm.sendDirectRAW(peerID, buff)
}

func (nm *NetMessenger) broadcastRAW(buff []byte, excs []string) error {
	nm.mutClosed.RLock()
	if nm.closed {
		nm.mutClosed.RUnlock()
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: "Attempt to write on a closed messenger!\n"}
	}
	nm.mutClosed.RUnlock()

	var errFound = &NodeError{}

	peers := nm.Peers()

	for i := 0; i < len(peers); i++ {
		peerID := peer.ID(peers[i]).Pretty()

		if peerID == nm.ID().Pretty() {
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

		err := nm.sendDirectRAW(peerID, buff)

		if err != nil {
			errNode, _ := err.(*NodeError)
			errFound.NestedErrors = append(errFound.NestedErrors, *errNode)
		}
	}

	if len(peers) <= 1 {
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: "Attempt to send to no one!\n"}
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

func (nm *NetMessenger) BroadcastBuff(buff []byte, excs []string) error {
	m := NewMessage(nm.ID().Pretty(), buff, nm.Marshalizer())

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: err.Error()}
	}

	return nm.broadcastRAW(buff, excs)
}

func (nm *NetMessenger) BroadcastString(message string, excs []string) error {
	return nm.BroadcastBuff([]byte(message), excs)
}

func (nm *NetMessenger) BroadcastMessage(m *Message, excs []string) error {
	if m == nil {
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: fmt.Sprintf("Can not broadcast NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: err.Error()}
	}

	return nm.broadcastRAW(buff, excs)
}

func (nm *NetMessenger) StreamHandler(stream libP2PNet.Stream) {
	peerID := stream.Conn().RemotePeer().Pretty()

	nm.mutChansSend.Lock()
	chanSend, ok := nm.chansSend[peerID]
	if !ok {
		chanSend = make(chan []byte, 10000)
		nm.chansSend[peerID] = chanSend
	}
	nm.mutChansSend.Unlock()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go func(rw *bufio.ReadWriter, queue *MessageQueue) {
		for {
			buff, _ := rw.ReadBytes(byte('\n'))

			nm.gotNewMessage(buff, stream.Conn().RemotePeer())
		}

	}(rw, nm.queue)

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

func (nm *NetMessenger) gotNewMessage(buff []byte, remotePeer peer.ID) {
	if len(buff) == 0 {
		return
	}

	if (len(buff) == 1) && (buff[0] == byte('\n')) {
		return
	}

	if buff[len(buff)-1] == byte('\n') {
		buff = buff[0 : len(buff)-1]
	}

	m, err := CreateFromByteArray(nm.Marshalizer(), buff)

	if err != nil {
		return
	}

	err = m.VerifyAndSetSigned()
	if err != nil {
		return
	}

	sha3 := crypto.SHA3_256.New()
	b64 := base64.StdEncoding
	hash := b64.EncodeToString(sha3.Sum([]byte(m.Payload)))

	if nm.queue.ContainsAndAdd(hash) {
		return
	}

	if nm.onMsgRecv != nil {
		nm.onMsgRecv(nm, remotePeer.Pretty(), m)
	}
}

func (nm *NetMessenger) Bootstrap(ctx context.Context) {
	nm.mutClosed.RLock()
	if nm.closed {
		nm.mutClosed.RUnlock()
		return
	}
	nm.mutClosed.RUnlock()

	nm.mutBootstrap.Lock()
	defer nm.mutBootstrap.Unlock()

	if nm.mdns != nil {
		return
	}

	nm.dn = NewDiscoveryNotifier(nm)

	mdns, err := discovery.NewMdnsService(context.Background(), nm.p2pNode, time.Second, "discovery")

	if err != nil {
		panic(err)
	}

	mdns.RegisterNotifee(nm.dn)
	nm.mdns = mdns

	wait := time.Second * 10
	fmt.Printf("\n**** Waiting %v to bootstrap...****\n\n", wait)

	nm.cn.Start()

	time.Sleep(wait)
}

func (nm *NetMessenger) PrintConnected() {
	conns := nm.Conns()

	fmt.Printf("Node %s is connected to: \n", nm.ID().Pretty())

	for i := 0; i < len(conns); i++ {
		fmt.Printf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(nm.ID(), conns[i].RemotePeer()))
	}
}

func (nm *NetMessenger) AddAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	nm.p2pNode.Network().Peerstore().AddAddr(p, addr, ttl)
}

func (nm *NetMessenger) Connectedness(pid peer.ID) libP2PNet.Connectedness {
	return nm.p2pNode.Network().Connectedness(pid)
}

func (nm *NetMessenger) ParseAddressIpfs(address string) (*pstore.PeerInfo, error) {
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
