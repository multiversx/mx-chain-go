package ex01ChatLibP2P

/*
*
* The MIT License (MIT)
*
* Copyright (c) 2014 Juan Batiz-Benet
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in
* all copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
* THE SOFTWARE.
*
* This program demonstrate a simple chat application using p2p communication.
*
 */

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"

	"github.com/libp2p/go-libp2p"

	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"strconv"
	"time"
)

var firstNodeHostID string
var portFirstNode int64

var startSend time.Time

/*
* addAddrToPeerstore parses a peer multiaddress and adds
* it to the given host's peerstore, so it knows how to
* contact it. It returns the peer ID of the remote peer.
* @credit examples/http-proxy/proxy.go
 */
func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	// The following code extracts target's the peer ID from the
	// given multiaddress
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

	// We have a peer ID and a targetAddr so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

func handleStreamSecondNode(s net.Stream) {
	log.Println("Second node: Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw, "second node")
	//go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func handleStreamFirstNode(s net.Stream) {
	log.Println("First node: Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw, "first node")

	// stream 's' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter, identifier string) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("%s got in %v> %s", identifier, time.Now().Sub(startSend), str)
		}

	}
}

func writeData(rw *bufio.ReadWriter, identifier string) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		//fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		fmt.Printf("Sending from %s > %s", identifier, sendData)
		startSend = time.Now()
		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}

}

func startFirstNode(port int64) {
	portFirstNode = port

	start := time.Now()

	// If debug is enabled used constant random source else cryptographic randomness.
	var r io.Reader
	// Constant random source. This will always generate the same host ID on multiple execution.
	// Don't do this in production code.
	r = mrand.New(mrand.NewSource(port))

	// Creates a new RSA key pair for this host
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

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

	firstNodeHostID = host.ID().Pretty()

	// Set a function as stream handler.
	// This function  is called when a peer initiate a connection and starts a stream with this peer.
	// Only applicable on the receiving side.
	host.SetStreamHandler("/chat/1.0.0", handleStreamFirstNode)

	fmt.Printf("First node: Connection took %v, waiting for incoming connection\n\n", time.Now().Sub(start))

	// Hang forever
	<-make(chan struct{})
}

func startSecondNode(port int64) {

	start := time.Now()
	// If debug is enabled used constant random source else cryptographic randomness.
	var r io.Reader
	// Constant random source. This will always generate the same host ID on multiple execution.
	// Don't do this in production code.
	r = mrand.New(mrand.NewSource(port))

	// Creates a new RSA key pair for this host
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

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

	for firstNodeHostID == "" {
		//fmt.Println("Waiting for first node...")
		time.Sleep(time.Microsecond)
	}

	dest := "/ip4/127.0.0.1/tcp/" + strconv.Itoa(int(portFirstNode)) + "/ipfs/" + firstNodeHostID

	// Add destination peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	peerID := addAddrToPeerstore(host, dest)

	fmt.Println("This node's multiaddress: ")
	// IP will be 0.0.0.0 (listen on any interface) and port will be 0 (choose one for me).
	// Although this node will not listen for any connection. It will just initiate a connect with
	// one of its peer and use that stream to communicate.
	fmt.Printf("%s/ipfs/%s\n", sourceMultiAddr, host.ID().Pretty())

	// Start a stream with peer with peer Id: 'peerId'.
	// Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
	s, err := host.NewStream(context.Background(), peerID, "/chat/1.0.0")

	if err != nil {
		panic(err)
	}

	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	fmt.Printf("Second node: Connection took %v\n\n", time.Now().Sub(start))

	// Create a thread to read and write data.
	go writeData(rw, "second node")
	go readData(rw, "second node")

	// Hang forever.
	select {}
}

func Main() {
	go startFirstNode(4000)
	go startSecondNode(4001)

	// Hang forever.
	select {}
}
