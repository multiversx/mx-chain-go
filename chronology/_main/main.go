package main

import (
	"context"
	"crypto"
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/synctime"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/statistic"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-peer"

	"sync"
	"time"
)

var CONSENSUS_GROUP_SIZE *int
var FIRST_NODE_ID *int
var LAST_NODE_ID *int
var SYNC_MODE *bool
var MAX_ALLOWED_PEERS *int
var GENESIS_TIME_STAMP = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
var ROUND_DURATION *time.Duration

func main() {
	// Parse options from the command line
	CONSENSUS_GROUP_SIZE = flag.Int("size", 21, "consensus group size which validate proposed block by the leader")
	MAX_ALLOWED_PEERS = flag.Int("peers", 4, "max connections allowed by each peer")
	ROUND_DURATION = flag.Duration("duration", 4000*time.Millisecond, "round duration in milliseconds")
	FIRST_NODE_ID = flag.Int("first", 1, "first node ID. This ID should be between 1 and consensus group size")
	LAST_NODE_ID = flag.Int("last", 21, "last node ID. This ID should be between 1 and consensus group size, but also greater or equal than first node ID")
	SYNC_MODE = flag.Bool("sync", false, "sync mode in subrounds will be used")
	flag.Parse()

	if *FIRST_NODE_ID < 1 || *LAST_NODE_ID > *CONSENSUS_GROUP_SIZE || *CONSENSUS_GROUP_SIZE < 1 || *FIRST_NODE_ID > *LAST_NODE_ID || *MAX_ALLOWED_PEERS < 1 || *MAX_ALLOWED_PEERS > *CONSENSUS_GROUP_SIZE-1 {
		fmt.Println("Eroare parametrii de intrare")
		return
	} else {
		fmt.Printf("size = %d\npeers = %d\nduration = %d\nfirst = %d\nlast = %d\nsynctime = %v\n\n", *CONSENSUS_GROUP_SIZE, *MAX_ALLOWED_PEERS, *ROUND_DURATION, *FIRST_NODE_ID, *LAST_NODE_ID, *SYNC_MODE)
	}

	marsh := &mock.MockMarshalizer{}

	var nodes []*p2p.Messenger

	for i := *FIRST_NODE_ID; i <= *LAST_NODE_ID; i++ {
		node, err := createMessenger(Network, 4000+i-1, *MAX_ALLOWED_PEERS, marsh)

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

	consensusGroup := make([]string, 0)

	for i := 1; i <= *CONSENSUS_GROUP_SIZE; i++ {
		consensusGroup = append(consensusGroup, p2p.NewConnectParamsFromPort(4000+i-1).ID.Pretty())
	}

	vld := validators.New(consensusGroup)

	PBFTThreshold := len(vld.ConsensusGroup)*2/3 + 1
	cns := consensus.New(consensus.Threshold{1, PBFTThreshold, PBFTThreshold, PBFTThreshold, PBFTThreshold})

	sync := synctime.New(*ROUND_DURATION)

	var chrs []*chronology.Chronology

	for i := 0; i < len(nodes); i++ {
		vld.Self = consensusGroup[*FIRST_NODE_ID+i-1]
		bc := blockchain.New(nil, &sync, i == 0)
		stats := statistic.New()

		chr := chronology.New(nodes[i], &vld, &cns, &stats, &sync, &bc, GENESIS_TIME_STAMP, *ROUND_DURATION)
		chr.DoLog = i == 0
		chr.DoSyncMode = *SYNC_MODE
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

			currentBlock := chrs[i].BlockChain.GetCurrentBlock()
			if currentBlock == nil {
				continue
			}

			if currentBlock.GetNonce() > oldNounce[i] {
				oldNounce[i] = currentBlock.GetNonce()
				spew.Dump(currentBlock)
				fmt.Printf("\n********** There was %d rounds and was proposed %d blocks, which means %.2f%% hit rate **********\n", chrs[i].Stats.GetRounds(), chrs[i].Stats.GetRoundsWithBlock(), float64(chrs[i].Stats.GetRoundsWithBlock())*100/float64(chrs[i].Stats.GetRounds()))
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
		//		cp := p2p.NewConnectParamsFromPortAndIP(port, GetOutboundIP())
		cp := p2p.NewConnectParamsFromPort(port)

		return p2p.NewNetMessenger(context.Background(), marsh, *cp, []string{}, maxAllowedPeers)
	default:
		panic("Type not defined!")
	}
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
