package main

import (
	"context"
	"flag"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/validators"
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
var GENESIS_TIME_STAMP = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
var ROUND_DURATION time.Duration
var ID int
var TEST_SINGLE_INSTANCE = false

func main() {

	bcs := data.GetBlockChainerService()

	// Parse options from the command line
	NODES = *flag.Int("nodes", 21, "consensus group size")
	ID = *flag.Int("id", 0, "node name")
	START_PORT = *flag.Int("start_port", 4000, "port")
	MAX_CONNECTION_PEERS = *flag.Int("max_connection_peers", 4, "max connections with peers")
	ROUND_DURATION = *flag.Duration("round_duration", 600*time.Millisecond, "round duration")
	flag.Parse()

	if TEST_SINGLE_INSTANCE {

		node, err := p2p.NewNode(context.Background(), START_PORT+ID, []string{}, service.GetMarshalizerService(), MAX_CONNECTION_PEERS)

		if err != nil {
			spew.Printf("Error NewNode\n")
			return
		}

		cp := p2p.NewClusterParameter("127.0.0.1", START_PORT, START_PORT+NODES-1)

		time.Sleep(time.Second)

		node.Bootstrap(context.Background(), []p2p.ClusterParameter{*cp})

		time.Sleep(time.Second)

		conns := node.P2pNode.Network().Conns()

		spew.Printf("Node %s is connected to: \n", node.P2pNode.ID().Pretty())

		for i := 0; i < len(conns); i++ {
			spew.Printf("\t- %s\n", conns[i].RemotePeer().Pretty())
		}

		spew.Printf("\n\n")

		consensusGroup := make([]string, NODES)

		for i := 0; i < NODES; i++ {
			consensusGroup[i] = string(i)
		}

		v := validators.Validators{}
		v.SetConsensusGroup(consensusGroup)

		v.SetSelf(string(ID))
		csi := consensus.NewConsensusServiceImpl(node, &v, GENESIS_TIME_STAMP, ROUND_DURATION)
		go csi.StartRounds()

		oldNounce := -1

		for {
			time.Sleep(100 * time.Millisecond)

			currentBlock := bcs.GetCurrentBlock(&csi.BlockChain)
			if currentBlock == nil {
				continue
			}

			if currentBlock.GetNonce() > oldNounce {
				oldNounce = currentBlock.GetNonce()
				spew.Dump(currentBlock)
			}
		}

		csi.DoRun = false
		node.P2pNode.Close()

	} else {
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

		consensusGroup := make([]string, NODES)

		for i := 0; i < NODES; i++ {
			consensusGroup[i] = nodes[i].P2pNode.ID().Pretty()
		}

		v := validators.Validators{}
		v.SetConsensusGroup(consensusGroup)

		var csis []*consensus.ConsensusServiceImpl

		for i := 0; i < NODES; i++ {
			v.SetSelf(consensusGroup[i])
			csi := consensus.NewConsensusServiceImpl(nodes[i], &v, GENESIS_TIME_STAMP, ROUND_DURATION)
			csi.DoLog = i == 0
			csis = append(csis, csi)
		}

		for i := 0; i < NODES; i++ {
			go csis[i].StartRounds()
		}

		var oldNounce []int

		for i := 0; i < NODES; i++ {
			oldNounce = append(oldNounce, -1)
		}

		for {
			time.Sleep(100 * time.Millisecond)

			for i := 0; i < NODES; i++ {
				if !csis[i].DoLog {
					continue
				}

				currentBlock := bcs.GetCurrentBlock(&csis[i].BlockChain)
				if currentBlock == nil {
					continue
				}

				if currentBlock.GetNonce() > oldNounce[i] {
					oldNounce[i] = currentBlock.GetNonce()
					spew.Dump(currentBlock)
				}
			}
		}

		for i := 0; i < NODES; i++ {
			csis[i].DoRun = false
			nodes[i].P2pNode.Close()
		}
	}
}
