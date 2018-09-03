package ex02AdvChatLibP2P

import (
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"time"
)

var isFirstNode bool

//var port int = 4000
var remoteConnection string

func extractPeerInfo(h host.Host, addr string) (peerstore.PeerInfo, peer.ID) {
	ipfsaddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := ipfsaddr.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := multiaddr.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	return peerstore.PeerInfo{peerid, []ma.Multiaddr{targetAddr}}, peerid
}

/*
* addAddrToPeerstore parses a peer multiaddress and adds
* it to the given host's peerstore, so it knows how to
* contact it. It returns the peer ID of the remote peer.
* @credit examples/http-proxy/proxy.go
 */
//func addAddrToPeerstore(h host.Host, addr string) peer.ID {
//	// The following code extracts target's the peer ID from the
//	// given multiaddress
//	ipfsaddr, err := multiaddr.NewMultiaddr(addr)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	pid, err := ipfsaddr.ValueForProtocol(multiaddr.P_IPFS)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	peerid, err := peer.IDB58Decode(pid)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	// Decapsulate the /ipfs/<peerID> part from the target
//	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
//	targetPeerAddr, _ := multiaddr.NewMultiaddr(
//		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
//	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
//
//	// We have a peer ID and a targetAddr so we add
//	// it to the peerstore so LibP2P knows how to contact it
//	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
//	return peerid
//}

func prepareOpt() {
	stdReader := bufio.NewReader(os.Stdin)

	var entry01 int

	fmt.Println("Advanced chat v.1.0.0.0")
	fmt.Println("=======================")

	for entry01 == 0 {
		fmt.Println()
		fmt.Println("Choose option:")
		fmt.Println(" 1. Start as first node")
		fmt.Println(" 2. Connect to a node")
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		filtered := ""

		for _, runeValue := range sendData {
			if runeValue == 13 {
				continue
			}

			if runeValue == 10 {
				continue
			}

			filtered = filtered + string(runeValue)
		}

		switch filtered {
		case "1":
			entry01 = 1
			break
		case "2":
			entry01 = 2
			break
		default:
			fmt.Println("Please choose a valid entry!")
		}
	}

	isFirstNode = (entry01 == 1)

	if isFirstNode {
		return
	}

	fmt.Println("Enter destination in '/ip4/192.168.0.1/tcp/4000/ipfs/QmdSyhb8eR9dDSR5jjnRoTDBwpBCSAjT7WueKJ9cQArYoA' fomat:")

	sendData, err := stdReader.ReadString('\n')

	if err != nil {
		panic(err)
	}

	filtered := ""

	for _, runeValue := range sendData {
		if runeValue == 13 {
			continue
		}

		if runeValue == 10 {
			continue
		}

		filtered = filtered + string(runeValue)
	}

	remoteConnection = filtered
}

func handleStream(s net.Stream) {
	log.Println("First node: Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)
	// stream 's' will stay open until you close it (or the other side closes it).
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		//fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}

}

func connHandler(conn net.Conn) {
	if conn == nil {
		panic("nil conn")
	}

	fmt.Printf("connection from %s, status %v", conn.RemotePeer().Pretty(), conn.Stat())
}

func connectAndRun(remoteConn *string) {
	start := time.Now()

	// If debug is enabled used constant random source else cryptographic randomness.
	var r io.Reader
	// Constant random source. This will always generate the same host ID on multiple execution.
	// Don't do this in production code.
	r = mrand.New(mrand.NewSource(time.Now().Unix()))

	// Creates a new RSA key pair for this host
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	var localPort = 4000

	if !isFirstNode {
		localPort = 4000 + mrand.New(mrand.NewSource(time.Now().Unix())).Intn(1000)
	}

	// 0.0.0.0 will listen on any interface device
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", localPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.

	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Current node is: /ip4/127.0.0.1/tcp/%d/ipfs/%s\n", localPort, host.ID().Pretty())
	fmt.Println(" You can replace 127.0.0.1 with a public IP")

	host.Network().SetConnHandler(connHandler)

	if isFirstNode {
		// Set a function as stream handler.
		// This function  is called when a peer initiate a connection and starts a stream with this peer.
		// Only applicable on the receiving side.
		host.SetStreamHandler("/chat/1.0.0.0", handleStream)
	} else {
		peerInfo, _ := extractPeerInfo(host, *remoteConn)

		host.SetStreamHandler("/chat/1.0.0.0", handleStream)

		errConnect := host.Connect(context.Background(), peerInfo)

		if errConnect != nil {
			panic(errConnect)
		}

		//for pid := range host.Network().Peers(){
		//
		//
		//
		//}

		//
		//// Start a stream with peer with peer Id: 'peerId'.
		//// Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
		//s, err := host.NewStream(context.Background(), peerID, "/chat/1.0.0.0")
		//
		//if err != nil {
		//	panic(err)
		//}
		//
		//// Create a buffered stream so that read and writes are non blocking.
		//rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		//
		//// Create a thread to read and write data.
		//go writeData(rw)
		//go readData(rw)

	}

	fmt.Printf("Connection took %v, waiting for incoming connection\n\n", time.Now().Sub(start))

	select {}
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("Recv > %s", str)
		}
	}
}

func Main() {
	prepareOpt()

	if isFirstNode {
		connectAndRun(nil)
	} else {
		connectAndRun(&remoteConnection)
	}
}
