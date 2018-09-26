package main

import (
	"context"
	"flag"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"github.com/davecgh/go-spew/spew"
	"time"
)

var NODES int
var START_PORT int
var MAX_CONNECTION_PEERS int
var GENESIS_TIME_STAMP = time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)
var ROUND_DURATION time.Duration

func main() {

	// Parse options from the command line
	NODES = *flag.Int("nodes", 2, "consensus group size")
	START_PORT = *flag.Int("start_port", 4000, "round time in seconds")
	MAX_CONNECTION_PEERS = *flag.Int("max_connection_peers", 2, "node id")
	ROUND_DURATION = *flag.Duration("round_duration", 4*time.Second, "node id")
	flag.Parse()

	var nodes []*p2p.Node

	for i := 0; i < NODES; i++ {
		node, err := p2p.NewNode(context.Background(), START_PORT+i, []string{}, service.GetMarshalizerService(), MAX_CONNECTION_PEERS)

		if err != nil {
			spew.Printf("Error NewNode\n")
			return
		}

		nodes = append(nodes, node)
	}

	cp := p2p.NewClusterParameter("127.0.0.1", START_PORT, START_PORT+NODES-1)

	time.Sleep(time.Second)

	for i := 0; i < NODES; i++ {
		nodes[i].Bootstrap(context.Background(), []p2p.ClusterParameter{*cp})
	}

	time.Sleep(time.Second)

	for i := 0; i < NODES; i++ {
		conns := nodes[i].P2pNode.Network().Conns()

		spew.Printf("Node %s is connected to: \n", nodes[i].P2pNode.ID().Pretty())

		for j := 0; j < len(conns); j++ {
			spew.Printf("\t- %s\n", conns[j].RemotePeer().Pretty())
		}
	}

	spew.Printf("\n\n")

	ch := make(chan string)

	csis := []*consensus.ConsensusServiceImpl{}

	for i := 0; i < NODES; i++ {
		csi := consensus.NewConsensusServiceImpl(nodes, i, GENESIS_TIME_STAMP, ROUND_DURATION, &ch)
		csis = append(csis, &csi)
	}

	for i := 0; i < NODES; i++ {
		go csis[i].StartRounds()
	}

	var oldNounce []int

	for i := 0; i < NODES; i++ {
		oldNounce = append(oldNounce, -1)
	}

	for v := range ch {
		spew.Printf(v + "\n")
		for i := 0; i < NODES; i++ {
			currentBlock := data.BlockChainServiceImpl{}.GetCurrentBlock(&csis[i].BlockChain)
			if currentBlock == nil {
				continue
			}

			if currentBlock.GetNonce() > oldNounce[i] {
				oldNounce[i] = currentBlock.GetNonce()
				spew.Printf("\nNode with ID = " + csis[i].P2PNode.P2pNode.ID().Pretty() + " has added a new block in his own blockchain\n\n")
				spew.Dump(currentBlock)
				spew.Printf("\n")
			}
		}
	}

	for i := 0; i < NODES; i++ {
		csis[i].DoRunning = false
		nodes[i].P2pNode.Close()
	}
}
