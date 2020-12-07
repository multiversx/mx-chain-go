package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

func createMockNetworkArgs() libp2p.ArgsNetworkMessenger {
	return libp2p.ArgsNetworkMessenger{
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:                          true,
				Type:                             "optimized",
				RefreshIntervalInSec:             2,
				ProtocolID:                       "/erd/kad/1.0.0",
				InitialPeerList:                  nil,
				BucketSize:                       100,
				RoutingTableRefreshIntervalInSec: 10,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer: &libp2p.LocalSyncTimer{},
	}
}

//The purpose of this example program is to show what happens if a peer connects to a network of 100 peers
func main() {
	startingPort := 32000

	advertiser, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())
	startingPort++
	fmt.Printf("advertiser is %s\n", getConnectableAddress(advertiser))
	peers := make([]p2p.Messenger, 0)
	_ = advertiser.Bootstrap(0)

	for i := 0; i < 99; i++ {
		arg := createMockNetworkArgs()
		arg.P2pConfig.KadDhtPeerDiscovery.InitialPeerList = []string{getConnectableAddress(advertiser)}
		netPeer, _ := libp2p.NewNetworkMessenger(arg)
		_ = netPeer.Bootstrap(0)

		peers = append(peers, netPeer)

		startingPort++
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
