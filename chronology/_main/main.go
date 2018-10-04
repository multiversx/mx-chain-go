package main

import (
	"context"
	"crypto"
	"flag"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-peer"

	"sync"
	"time"
)

var NODES int
var START_PORT int
var END_PORT int
var MAX_ALLOWED_PEERS int
var GENESIS_TIME_STAMP = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
var ROUND_DURATION time.Duration
var ID int

func main() {

	bcs := data.GetBlockChainerService()

	// Parse options from the command line
	NODES = *flag.Int("nodes", 21, "consensus group size")
	START_PORT = *flag.Int("start_port", 4000, "start port")
	END_PORT = *flag.Int("end_port", 4020, "end port")
	MAX_ALLOWED_PEERS = *flag.Int("max_allowed_peers", 4, "max allowed peers")
	ROUND_DURATION = *flag.Duration("round_duration", 8000*time.Millisecond, "round duration")
	flag.Parse()

	marsh := &mock.MockMarshalizer{}

	var nodes []*p2p.Messenger

	for i := START_PORT; i <= END_PORT; i++ {
		node, err := createMessenger(Network, i, MAX_ALLOWED_PEERS, marsh)

		if err != nil {
			continue
		}

		nodes = append(nodes, &node)
	}

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]

		go func() {
			(*node).Bootstrap(context.Background())
			wg.Done()
		}()
	}

	wg.Wait()

	for i := 0; i < len(nodes); i++ {
		(*nodes[i]).PrintConnected()
		spew.Println()
	}

	time.Sleep(time.Second)

	/*
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
	*/
	consensusGroup := make([]string, 0)

	for i := 0; i < len(nodes); i++ {
		consensusGroup = append(consensusGroup, (*nodes[i]).ID().Pretty())
	}

	v := validators.Validators{}
	v.SetConsensusGroup(consensusGroup)

	var chrs []*chronology.Chronology

	for i := 0; i < len(nodes); i++ {
		v.SetSelf(consensusGroup[i])
		chr := chronology.New(nodes[i], &v, GENESIS_TIME_STAMP, ROUND_DURATION)
		chr.DoLog = i == 0
		chr.DoSyncMode = true
		chrs = append(chrs, chr)
	}

	for i := 0; i < len(nodes); i++ {
		go chrs[i].StartRounds()
	}

	var oldNounce []int

	for i := 0; i < len(nodes); i++ {
		oldNounce = append(oldNounce, -1)
	}

	for {
		time.Sleep(100 * time.Millisecond)

		for i := 0; i < len(nodes); i++ {
			if i > 0 {
				continue
			}

			currentBlock := bcs.GetCurrentBlock(&chrs[i].BlockChain)
			if currentBlock == nil {
				continue
			}

			if currentBlock.GetNonce() > oldNounce[i] {
				oldNounce[i] = currentBlock.GetNonce()
				spew.Dump(currentBlock)
			}
		}
	}

	for i := 0; i < len(nodes); i++ {
		chrs[i].DoRun = false
		(*nodes[i]).Close()
	}
}

const (
	Memory int = iota
	Network
)

func createMessenger(mesType int, port int, maxAllowedPeers int, marsh marshal.Marshalizer) (p2p.Messenger, error) {
	switch mesType {
	case Memory:
		sha3 := crypto.SHA3_256.New()
		id := peer.ID(sha3.Sum([]byte("Name" + strconv.Itoa(port))))

		return p2p.NewMemoryMessenger(marsh, id, maxAllowedPeers)
	case Network:
		cp := p2p.NewConnectParamsFromPort(port)

		return p2p.NewNetMessenger(context.Background(), marsh, *cp, []string{}, maxAllowedPeers)
	default:
		panic("Type not defined!")
	}
}
