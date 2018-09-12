package p2p

import (
	"bufio"
	"context"
	"crypto"
	"encoding/base64"
	"fmt"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-floodsub"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"strings"

	//"github.com/libp2p/go-libp2p"
	//"github.com/libp2p/go-libp2p-crypto"
	//"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-metrics"
	libP2PNet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peerstore"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	tu2 "github.com/libp2p/go-libp2p-swarm/testing"
	//"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-tcp-transport"
	"time"
	//mrand "math/rand"
)

type Node struct {
	P2pNode   host.Host
	chansSend map[string]chan string

	queue *MessageQueue

	OnMsgRecvBroadcast func(sender *Node, topic string, msg *floodsub.Message)
	OnMsgRecv          func(sender *Node, peerID string, message string)
}

func GenSwarm(ctx context.Context, port int) *swarm.Swarm {
	p := randPeerNetParamsOrFatal(port)

	ps := pstore.NewPeerstore()
	ps.AddPubKey(p.ID, p.PubKey)
	ps.AddPrivKey(p.ID, p.PrivKey)
	s := swarm.NewSwarm(ctx, p.ID, ps, metrics.NewBandwidthCounter())

	tcpTransport := tcp.NewTCPTransport(tu2.GenUpgrader(s))
	tcpTransport.DisableReuseport = false

	if err := s.AddTransport(tcpTransport); err != nil {
		panic(err)
	}

	if err := s.Listen(p.Addr); err != nil {
		panic(err)
	}

	s.Peerstore().AddAddrs(p.ID, s.ListenAddresses(), pstore.PermanentAddrTTL)

	return s
}

func CreateNewNode(ctx context.Context, port int, addresses []string) *Node {
	var node Node

	timeStart := time.Now()

	netw := GenSwarm(ctx, port)
	h := basichost.New(netw)

	fmt.Printf("Node: %v has the following addr table: \n", h.ID().Pretty())

	for i, addr := range h.Addrs() {
		fmt.Printf("%d: %s/ipfs/%s\n", i, addr, h.ID().Pretty())
	}
	fmt.Println()

	fmt.Printf("Created node in %v\n", time.Now().Sub(timeStart))

	node.P2pNode = h
	node.chansSend = make(map[string](chan string))
	node.queue = NewMessageQueue(50000)

	node.addRWHandlers("benchmark/nolimit/1.0.0.0")

	node.ConnectToAddresses(ctx, addresses)

	return &node
}

func (node *Node) ConnectToAddresses(ctx context.Context, addresses []string) {
	peers := 0

	timeStart := time.Now()

	for _, address := range addresses {
		addr, err := ipfsaddr.ParseString(address)
		if err != nil {
			panic(err)
		}
		pinfo, _ := peerstore.InfoFromP2pAddr(addr.Multiaddr())

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

func (node *Node) SendDirect(peerID string, message string) {
	chanSend, ok := node.chansSend[peerID]
	if !ok {
		fmt.Printf("Can not send to %v. Not connected?\n", peerID)
		return
	}

	select {
	case chanSend <- message:
	default:
		fmt.Printf("Can not send to %v. Pipe full! Message discarded!\n", peerID)
	}
}

func (node *Node) Broadcast(message string, excs []string) {
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

		node.SendDirect(peerID, message)
	}
}

func (node *Node) streamHandler(stream libP2PNet.Stream) {
	peerID := stream.Conn().RemotePeer().Pretty()

	chanSend, ok := node.chansSend[peerID]
	if !ok {
		chanSend = make(chan string, 10000)
		node.chansSend[peerID] = chanSend
	}

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go func(rw *bufio.ReadWriter, queue *MessageQueue) {
		for {
			str, _ := rw.ReadString('\n')

			if str == "" {
				continue
			}

			if str != "\n" {
				str = strings.Trim(str, "\n")
				sha3 := crypto.SHA3_256.New()
				base64 := base64.StdEncoding
				hash := base64.EncodeToString(sha3.Sum([]byte(str)))

				if queue.Contains(hash) {
					continue
				}

				queue.Add(hash)

				if node.OnMsgRecv != nil {
					node.OnMsgRecv(node, stream.Conn().RemotePeer().Pretty(), str)
				}
			}
		}
	}(rw, node.queue)

	go func(rw *bufio.ReadWriter, chanSend chan string) {
		for {
			data := <-chanSend

			rw.WriteString(fmt.Sprintf("%s\n", data))
			rw.Flush()
		}
	}(rw, chanSend)
}

func (node *Node) addRWHandlers(protocolID protocol.ID) {
	node.P2pNode.SetStreamHandler(protocolID, node.streamHandler)
}
