package p2p

import (
	"context"
	"crypto"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/btcsuite/btcutil/base58"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

var mutGloballyRegPeers *sync.Mutex
var GloballyRegisteredPeers map[peer.ID]*MemoryMessenger

func init() {
	mutGloballyRegPeers = &sync.Mutex{}
	GloballyRegisteredPeers = make(map[peer.ID]*MemoryMessenger)
}

type MemoryMessenger struct {
	peerID peer.ID

	mutConnectedPeers sync.Mutex
	connectedPeers    map[peer.ID]*MemoryMessenger

	marsh marshal.Marshalizer
	rt    *RoutingTable
	queue *MessageQueue

	maxAllowedPeers int

	onMsgRecv func(caller Messenger, peerID string, m *Message)

	mutClosed sync.RWMutex
	closed    bool
}

func NewMemoryMessenger(m marshal.Marshalizer, pid peer.ID, maxAllowedPeers int) (*MemoryMessenger, error) {
	if m == nil {
		return nil, errors.New("Marshalizer is nil! Can't create messenger!")
	}

	mm := MemoryMessenger{peerID: pid, marsh: m, maxAllowedPeers: maxAllowedPeers, closed: false}

	mm.mutConnectedPeers = sync.Mutex{}
	mm.mutConnectedPeers.Lock()
	mm.connectedPeers = make(map[peer.ID]*MemoryMessenger)
	mm.mutConnectedPeers.Unlock()

	mm.rt = NewRoutingTable(mm.peerID)
	mm.queue = NewMessageQueue(50000)

	mm.mutClosed = sync.RWMutex{}

	mutGloballyRegPeers.Lock()
	GloballyRegisteredPeers[pid] = &mm
	mutGloballyRegPeers.Unlock()

	return &mm, nil
}

func (mm *MemoryMessenger) Close() error {
	mm.mutClosed.Lock()
	mm.closed = true
	mm.mutClosed.Unlock()
	return nil
}

func (mm *MemoryMessenger) ID() peer.ID {
	return mm.peerID
}

func (mm *MemoryMessenger) Peers() []peer.ID {
	peers := make([]peer.ID, 0)

	mm.mutConnectedPeers.Lock()
	for k := range mm.connectedPeers {
		peers = append(peers, k)
	}
	mm.mutConnectedPeers.Unlock()

	return peers
}

func (mm *MemoryMessenger) Conns() []net.Conn {
	conns := make([]net.Conn, 0)

	mm.mutConnectedPeers.Lock()
	for k := range mm.connectedPeers {
		c := &MockConn{localPeer: mm.peerID, remotePeer: k}
		conns = append(conns, c)
	}
	mm.mutConnectedPeers.Unlock()

	return conns
}

func (mm *MemoryMessenger) Marshalizer() marshal.Marshalizer {
	return mm.marsh
}

func (mm *MemoryMessenger) RouteTable() *RoutingTable {
	return mm.rt
}

func (mm *MemoryMessenger) Addrs() []string {
	//b := []byte(mm.peerID.Pretty())

	//return []multiaddr.Multiaddr{&MockMultiAddr{buff: b}}
	return []string{string(mm.peerID.Pretty())}
}

func (mm *MemoryMessenger) GetOnRecvMsg() func(caller Messenger, peerID string, m *Message) {
	return mm.onMsgRecv
}

func (mm *MemoryMessenger) SetOnRecvMsg(f func(caller Messenger, peerID string, m *Message)) {
	mm.onMsgRecv = f
}

func (mm *MemoryMessenger) ConnectToAddresses(ctx context.Context, addresses []string) {
	for i := 0; i < len(addresses); i++ {
		addr := peer.ID(base58.Decode(addresses[i]))

		mutGloballyRegPeers.Lock()
		val, ok := GloballyRegisteredPeers[addr]
		mutGloballyRegPeers.Unlock()

		if !ok {
			fmt.Printf("Bootstrapping the peer '%v' failed! [not found]\n", addresses[i])
			continue
		}

		if mm.peerID == addr {
			//won't add self
			continue
		}

		mm.mutConnectedPeers.Lock()
		//connect this to other peer
		mm.connectedPeers[addr] = val
		//connect other the other peer to this
		val.connectedPeers[mm.peerID] = mm
		mm.mutConnectedPeers.Unlock()
	}
}

func (mm *MemoryMessenger) gotNewMessage(sender *MemoryMessenger, buff []byte) {
	mm.mutClosed.RLock()
	if mm.closed {
		return
	}
	mm.mutClosed.RUnlock()

	if len(buff) == 0 {
		return
	}

	if (len(buff) == 1) && (buff[0] == byte('\n')) {
		return
	}

	if buff[len(buff)-1] == byte('\n') {
		buff = buff[0 : len(buff)-1]
	}

	m, err := CreateFromByteArray(mm.marsh, buff)

	if err != nil {
		return
	}

	sha3 := crypto.SHA3_256.New()
	b64 := base64.StdEncoding
	hash := b64.EncodeToString(sha3.Sum([]byte(m.Payload)))

	if mm.queue.ContainsAndAdd(hash) {
		return
	}

	if mm.onMsgRecv != nil {
		go mm.onMsgRecv(mm, sender.peerID.Pretty(), m)
	}
}

func (mm *MemoryMessenger) sendDirectRAW(peerID string, buff []byte) error {
	mm.mutClosed.RLock()
	if mm.closed {
		return &NodeError{PeerRecv: peerID, PeerSend: mm.ID().Pretty(), Err: "Attempt to write on a closed messenger!\n"}
	}
	mm.mutClosed.RUnlock()

	if peerID == mm.ID().Pretty() {
		//send to self allowed
		mm.gotNewMessage(mm, buff)
		return nil
	}

	mm.mutConnectedPeers.Lock()
	val, ok := mm.connectedPeers[peer.ID(base58.Decode(peerID))]
	mm.mutConnectedPeers.Unlock()

	if (!ok) || (val == nil) {
		return &NodeError{PeerRecv: peerID, PeerSend: mm.ID().Pretty(), Err: fmt.Sprintf("Can not send to %v. Not connected?\n", peerID)}
	}

	val.gotNewMessage(mm, buff)

	return nil
}

func (mm *MemoryMessenger) SendDirectBuff(peerID string, buff []byte) error {
	m := NewMessage(mm.ID().Pretty(), buff, mm.marsh)

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: mm.ID().Pretty(), Err: err.Error()}
	}

	return mm.sendDirectRAW(peerID, buff)
}

func (mm *MemoryMessenger) SendDirectString(peerID string, message string) error {
	return mm.SendDirectBuff(peerID, []byte(message))
}

func (mm *MemoryMessenger) SendDirectMessage(peerID string, m *Message) error {
	if m == nil {
		return &NodeError{PeerRecv: peerID, PeerSend: mm.ID().Pretty(), Err: fmt.Sprintf("Can not send NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: peerID, PeerSend: mm.ID().Pretty(), Err: err.Error()}
	}

	return mm.sendDirectRAW(peerID, buff)
}

func (mm *MemoryMessenger) broadcastRAW(buff []byte, excs []string) error {
	mm.mutClosed.RLock()
	if mm.closed {
		return &NodeError{PeerRecv: "", PeerSend: mm.ID().Pretty(), Err: "Attempt to write on a closed messenger!\n"}
	}
	mm.mutClosed.RUnlock()

	var errFound = &NodeError{}

	peers := mm.Peers()

	for i := 0; i < len(peers); i++ {
		peerID := peer.ID(peers[i]).Pretty()

		if peerID == mm.ID().Pretty() {
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

		err := mm.sendDirectRAW(peerID, buff)

		if err != nil {
			errNode, _ := err.(*NodeError)
			errFound.NestedErrors = append(errFound.NestedErrors, *errNode)
		}
	}

	if len(peers) == 0 {
		return &NodeError{PeerRecv: "", PeerSend: mm.ID().Pretty(), Err: "Attempt to send to no one!\n"}
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

func (mm *MemoryMessenger) BroadcastBuff(buff []byte, excs []string) error {
	m := NewMessage(mm.peerID.Pretty(), buff, mm.marsh)

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: mm.peerID.Pretty(), Err: err.Error()}
	}

	return mm.broadcastRAW(buff, excs)
}

func (mm *MemoryMessenger) BroadcastString(message string, excs []string) error {
	return mm.BroadcastBuff([]byte(message), excs)
}

func (mm *MemoryMessenger) BroadcastMessage(m *Message, excs []string) error {
	if m == nil {
		return &NodeError{PeerRecv: "", PeerSend: mm.ID().Pretty(), Err: fmt.Sprintf("Can not broadcast NIL message!\n")}
	}

	buff, err := m.ToByteArray()
	if err != nil {
		return &NodeError{PeerRecv: "", PeerSend: mm.ID().Pretty(), Err: err.Error()}
	}

	return mm.broadcastRAW(buff, excs)
}

func (mm *MemoryMessenger) StreamHandler(stream net.Stream) {
	panic("implement me")
}

func (mm *MemoryMessenger) Bootstrap(ctx context.Context) {
	go mm.doBootstrap()
}

func (mm *MemoryMessenger) doBootstrap() {
	for {
		mm.mutClosed.RLock()
		if mm.closed {
			mm.mutClosed.RUnlock()
			return
		}
		mm.mutClosed.RUnlock()

		mutGloballyRegPeers.Lock()
		for k, v := range GloballyRegisteredPeers {
			if !mm.rt.Has(k) {
				mm.rt.Update(k)

				mm.mutConnectedPeers.Lock()
				mm.connectedPeers[k] = v
				mm.mutConnectedPeers.Unlock()
			}
		}
		mutGloballyRegPeers.Unlock()

		time.Sleep(time.Second)
	}

}

func (mm *MemoryMessenger) PrintConnected() {
	conns := mm.Conns()

	fmt.Printf("Node %s is connected to: \n", mm.ID().Pretty())

	for i := 0; i < len(conns); i++ {
		fmt.Printf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(mm.ID(), conns[i].RemotePeer()))
	}
}

func (mm *MemoryMessenger) AddAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	mutGloballyRegPeers.Lock()
	val, ok := GloballyRegisteredPeers[p]
	mutGloballyRegPeers.Unlock()

	if !ok {
		val = nil
	}

	mm.mutConnectedPeers.Lock()
	mm.connectedPeers[p] = val
	mm.mutConnectedPeers.Unlock()
}

func (mm *MemoryMessenger) Connectedness(pid peer.ID) net.Connectedness {
	mm.mutConnectedPeers.Lock()
	_, ok := mm.connectedPeers[pid]
	mm.mutConnectedPeers.Unlock()

	if ok {
		return net.Connected
	} else {
		return net.NotConnected
	}

}
