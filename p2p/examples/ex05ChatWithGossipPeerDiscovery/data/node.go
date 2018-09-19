package data

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/multiformats/go-multiaddr"
	"io"
)

// Stream .
const Stream = "stream"

// ID .
const ID = "ID"

//IncomingCallback .
type IncomingCallback func(*Node) net.StreamHandler

// Node .
type Node struct {
	Address    multiaddr.Multiaddr
	ID         peer.ID
	PS         peerstore.Peerstore
	Port       *int
	Incoming   chan *Message
	OutgoingID int
	MS         Store
	host       *basichost.BasicHost
}

func panicGuard(err error) {
	if err != nil {
		panic(err)
	}
}

// NewNode ..
func NewNode(r io.Reader, sourcePort *int, cb func(*Node) net.StreamHandler) (node *Node) {
	node = new(Node)

	node.MS = make(Store)
	node.Incoming = make(chan *Message)
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

	return
}

func (node *Node) createHost(cb IncomingCallback) *basichost.BasicHost {

	swarm := swarm.NewSwarm(context.Background(), node.ID, node.PS, nil)

	host := basichost.New(swarm)
	host.SetStreamHandler("/chat/1.0.0", cb(node))

	fmt.Println("Incomming address:")
	fmt.Printf("/ip4/127.0.0.1/tcp/%d/ipfs/%s\n", *node.Port, host.ID().Pretty())
	fmt.Printf("\nWaiting for incoming connection\n\n")
	return host
}

// ConnectToDest .
func (node *Node) ConnectToDest(dest *string) *bufio.ReadWriter {
	if *dest == "" {
		return nil
	}

	peerID := addAddrToPeerstore(node.host, *dest)

	fmt.Println("This node's multiaddress: ")
	fmt.Printf("%s/ipfs/%s\n", node.Address, node.host.ID().Pretty())

	s, err := node.host.NewStream(context.Background(), peerID, "/chat/1.0.0")

	panicGuard(err)

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	node.PS.Put(s.Conn().RemotePeer(), Stream, rw)

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

// --- GO routines

//SendToPeers .
func (node *Node) SendToPeers(message *Message) {
	bytes, _ := json.Marshal(*message)
	node.MS[message.Key()] = message
	peers := node.PS.Peers()
	for _, peer := range peers {
		if peer != node.ID {
			remoteRW, _ := node.PS.Get(peer, Stream)
			rw := remoteRW.(*bufio.ReadWriter)
			_, err := rw.Write(bytes)
			rw.Flush()
			panicGuard(err)
		}
	}
}
