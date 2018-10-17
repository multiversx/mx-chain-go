package p2p

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
)

// NetMessenger implements a libP2P node with added functionality
type NetMessenger struct {
	protocol protocol.ID
	p2pNode  host.Host
	mdns     discovery.Service

	mutChansSend sync.RWMutex
	chansSend    map[string]chan []byte

	mutBootstrap sync.Mutex

	queue  *MessageQueue
	marsh  marshal.Marshalizer
	hasher hashing.Hasher
	rt     *RoutingTable
	cn     *ConnNotifier
	dn     *DiscoveryNotifier

	onMsgRecv func(caller Messenger, peerID string, m *Message)

	mutClosed sync.RWMutex
	closed    bool

	mutTopics sync.RWMutex
	topics    map[string]*Topic
}

// NewNetMessenger creates a new instance of NetMessenger.
func NewNetMessenger(ctx context.Context, marsh marshal.Marshalizer, hasher hashing.Hasher,
	cp *ConnectParams, maxAllowedPeers int) (*NetMessenger, error) {

	if marsh == nil {
		return nil, errors.New("marshalizer is nil! Can't create node")
	}

	if hasher == nil {
		return nil, errors.New("hasher is nil! Can't create node")
	}

	node := NetMessenger{
		marsh:  marsh,
		hasher: hasher,
		topics: make(map[string]*Topic, 0),
	}

	node.cn = NewConnNotifier(&node)
	node.cn.MaxAllowedPeers = maxAllowedPeers

	timeStart := time.Now()

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cp.Port)),
		libp2p.Identity(cp.PrivKey),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

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

	node.p2pNode.SetStreamHandler(node.protocol, node.streamHandler)

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
			str, err := node.p2pNode.NewStream(ctx, pinfo.ID, node.protocol)
			if err != nil {
				return err
			} else {
				node.streamHandler(str)
			}
		}

		return nil
	}

	return &node, nil
}

func (nm *NetMessenger) streamHandler(stream net.Stream) {
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

			nm.gotNewMessage(buff, stream.Conn().RemotePeer(), false)
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

func (nm *NetMessenger) gotNewMessage(buff []byte, remotePeer peer.ID, fromLoopback bool) {
	if len(buff) == 0 {
		return
	}

	if (len(buff) == 1) && (buff[0] == byte('\n')) {
		return
	}

	if buff[len(buff)-1] == byte('\n') {
		buff = buff[0 : len(buff)-1]
	}

	m, err := CreateFromByteArray(nm.marsh, buff)

	if err != nil {
		return
	}

	err = m.VerifyAndSetSigned()
	if err != nil {
		return
	}

	hash := string(nm.hasher.Compute(string(m.Payload)))

	if nm.queue.ContainsAndAdd(hash) {
		return
	}

	//check for topic and pass the message
	nm.mutTopics.RLock()
	defer nm.mutTopics.RUnlock()
	t, ok := nm.topics[m.Type]

	if !ok {
		//discard and do not broadcast
		return
	}

	//make a copy of received message so it can broadcast in parallel with the processing
	mes := *m

	if fromLoopback {
		//wil not resend on send

		mes.Hops = 0

		//process message
		t.NewMessageReceived(&mes)

		return
	}

	//process message
	err = t.NewMessageReceived(&mes)

	//in case of error we do not broadcast the message
	//in most cases is a corrupt/bad formatted message that could not have been
	//unmarshaled
	if err != nil {
		return
	}

	//async broadcast to others
	go func() {
		m.AddHop(nm.ID().Pretty())
		nm.broadcastMessage(m, []string{remotePeer.Pretty()})
	}()

}

// Closes a NetMessenger
func (nm *NetMessenger) Close() error {
	nm.mutClosed.Lock()
	nm.closed = true
	nm.mutClosed.Unlock()

	nm.p2pNode.Close()

	return nil
}

// ID returns the current id
func (nm *NetMessenger) ID() peer.ID {
	return nm.p2pNode.ID()
}

// Peers returns the connected peers list
func (nm *NetMessenger) Peers() []peer.ID {
	return nm.p2pNode.Peerstore().Peers()
}

// Conns return the connections made by this memory messenger
func (nm *NetMessenger) Conns() []net.Conn {
	return nm.p2pNode.Network().Conns()
}

// Marshalizer returns the used marshalizer object
func (nm *NetMessenger) Marshalizer() marshal.Marshalizer {
	return nm.marsh
}

// Hasher returns the used object for hashing data
func (nm *NetMessenger) Hasher() hashing.Hasher {
	return nm.hasher
}

// RouteTable will return the RoutingTable object
func (nm *NetMessenger) RouteTable() *RoutingTable {
	return nm.rt
}

// Addrs will return all addresses bind to current messenger
func (nm *NetMessenger) Addrs() []string {
	addrs := make([]string, 0)

	for _, adrs := range nm.p2pNode.Addrs() {
		addrs = append(addrs, adrs.String()+"/ipfs/"+nm.ID().Pretty())
	}

	return addrs
}

// ConnectToAddresses is used to explicitly connect to a well known set of addresses
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

		str, err := nm.p2pNode.NewStream(ctx, pinfo.ID, nm.protocol)
		if err != nil {
			fmt.Printf("Streaming the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		nm.streamHandler(str)
		peers++
	}

	fmt.Printf("Connected to %d peers in %v\n", peers, time.Now().Sub(timeStart))
}

// Bootstrap will try to connect to as many peers as possible
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

	nm.cn.Start()
}

// PrintConnected displays the connected peers
func (nm *NetMessenger) PrintConnected() {
	conns := nm.Conns()

	fmt.Printf("Node %s is connected to: \n", nm.ID().Pretty())

	for i := 0; i < len(conns); i++ {
		fmt.Printf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(nm.ID(), conns[i].RemotePeer()))
	}
}

// AddAddr adds a new address to peer store
func (nm *NetMessenger) AddAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	nm.p2pNode.Network().Peerstore().AddAddr(p, addr, ttl)
}

// Connectedness tests for a connection between self and another peer
func (nm *NetMessenger) Connectedness(pid peer.ID) net.Connectedness {
	return nm.p2pNode.Network().Connectedness(pid)
}

// ParseAddressIpfs translates the string containing the address of the node to a PeerInfo object
func (nm *NetMessenger) ParseAddressIpfs(address string) (*peerstore.PeerInfo, error) {
	addr, err := ipfsaddr.ParseString(address)
	if err != nil {
		return nil, err
	}

	pinfo, err := peerstore.InfoFromP2pAddr(addr.Multiaddr())
	if err != nil {
		return nil, err
	}

	return pinfo, nil
}

// AddTopic registers a new topic to this messenger
func (nm *NetMessenger) AddTopic(t *Topic) error {
	nm.mutTopics.Lock()
	defer nm.mutTopics.Unlock()

	if t == nil {
		return errors.New("topic can not be nil")
	}

	_, ok := nm.topics[t.Name]

	if ok {
		return errors.New("topic already exists")
	}

	nm.topics[t.Name] = t
	t.OnNeedToSendMessage = func(mes *Message, flagSign bool) error {
		mes.AddHop(nm.ID().Pretty())

		// optionally sign the message
		if flagSign {
			err := mes.Sign(nm.p2pNode.Peerstore().PrivKey(nm.ID()))
			if err != nil {
				return err
			}
		}

		return nm.broadcastMessage(mes, []string{})
	}

	return nil
}

// GetTopic returns the topic from its name or nil if no topic with that name
// was ever registered
func (nm *NetMessenger) GetTopic(topicName string) *Topic {
	nm.mutTopics.RLock()
	defer nm.mutTopics.RUnlock()

	t, ok := nm.topics[topicName]

	if !ok {
		return nil
	}

	return t
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
		nm.gotNewMessage(buff, nm.ID(), true)

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

func (nm *NetMessenger) sendDirectMessage(peerID string, m *Message) error {
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
			//loopback
			nm.gotNewMessage(buff, nm.ID(), true)
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

func (nm *NetMessenger) broadcastMessage(m *Message, excs []string) error {
	if m == nil {
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: fmt.Sprintf("Can not broadcast NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: nm.ID().Pretty(), Err: err.Error()}
	}

	return nm.broadcastRAW(buff, excs)
}
