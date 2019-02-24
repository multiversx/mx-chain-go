package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	cr "github.com/libp2p/go-libp2p-crypto"
)

var r *rand.Rand

//The purpose of this example program is to show what happens if a peer connects to a network of 100 peers
func main() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	startingPort := 32000

	advertiser, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		startingPort,
		genPrivKey(),
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		p2p.PeerDiscoveryKadDht,
	)
	startingPort++
	fmt.Printf("advertiser is %s\n", getConnectableAddress(advertiser))
	peers := make([]p2p.Messenger, 0)
	go func() {
		_ = advertiser.KadDhtDiscoverNewPeers()
		time.Sleep(time.Second)
	}()

	for i := 0; i < 99; i++ {
		netPeer, _ := libp2p.NewNetworkMessenger(
			context.Background(),
			startingPort,
			genPrivKey(),
			nil,
			loadBalancer.NewOutgoingPipeLoadBalancer(),
			p2p.PeerDiscoveryKadDht,
		)
		startingPort++

		fmt.Printf("%s connecting to %s...\n",
			getConnectableAddress(netPeer),
			getConnectableAddress(advertiser))

		_ = netPeer.ConnectToPeer(getConnectableAddress(advertiser))
		_ = netPeer.KadDhtDiscoverNewPeers()

		peers = append(peers, netPeer)

		go func() {
			_ = netPeer.KadDhtDiscoverNewPeers()
			time.Sleep(time.Second)
		}()
	}

	//display func
	go func() {
		for {
			time.Sleep(time.Second)
			showConnections(advertiser, peers)
		}
	}()

	time.Sleep(time.Second * 15)

	_ = advertiser.Close()
	for _, peer := range peers {
		if peer == nil {
			continue
		}

		_ = peer.Close()
	}
}

func getConnectableAddress(peer p2p.Messenger) string {
	for _, adr := range peer.Addresses() {
		if strings.Contains(adr, "127.0.0.1") {
			return adr
		}
	}

	return ""
}

func genPrivKey() cr.PrivKey {
	prv, _, _ := cr.GenerateKeyPairWithReader(cr.Ed25519, 0, r)
	return prv
}

func showConnections(advertiser p2p.Messenger, peers []p2p.Messenger) {
	header := []string{"Node", "Address", "No. of conns"}

	lines := make([]*display.LineData, 0)
	lines = append(lines, createDataLine(advertiser, advertiser, peers))

	for i := 0; i < len(peers); i++ {
		lines = append(lines, createDataLine(peers[i], advertiser, peers))
	}

	table, _ := display.CreateTableString(header, lines)

	fmt.Println(table)
}

func createDataLine(peer p2p.Messenger, advertiser p2p.Messenger, peers []p2p.Messenger) *display.LineData {
	ld := &display.LineData{}

	if peer == nil {
		ld.Values = []string{"<NIL>", "<NIL>", "0"}
		return ld
	}

	nodeName := "Peer"
	if advertiser == peer {
		nodeName = "Advertiser"
	}

	ld.Values = []string{nodeName,
		getConnectableAddress(peer),
		strconv.Itoa(computeConnectionsCount(peer, advertiser, peers))}

	return ld
}

func computeConnectionsCount(peer p2p.Messenger, advertiser p2p.Messenger, peers []p2p.Messenger) int {
	if peer == nil {
		return 0
	}

	knownPeers := 0
	if peer.IsConnected(advertiser.ID()) {
		knownPeers++
	}

	for i := 0; i < len(peers); i++ {
		p := peers[i]

		if p == nil {
			continue
		}

		if peer.IsConnected(peers[i].ID()) {
			knownPeers++
		}
	}

	return knownPeers
}
