package main

import (
	"context"
	"crypto"
	"flag"
	"fmt"
	"net"
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

var CGS int
var START_PORT int
var END_PORT int
var SYNC bool
var MAX_ALLOWED_PEERS int
var GENESIS_TIME_STAMP = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
var ROUND_DURATION time.Duration
var ID int

func main() {

	bcs := data.GetBlockChainerService()

	// Parse options from the command line
	CGS = *flag.Int("size", 21, "consensus group size")
	MAX_ALLOWED_PEERS = *flag.Int("peers", 4, "max allowed peers")
	ROUND_DURATION = *flag.Duration("duration", 4000*time.Millisecond, "round duration in milliseconds")
	START_PORT = *flag.Int("start_port", 4000, "start port")
	END_PORT = *flag.Int("end_port", 4015, "end port")
	SYNC = *flag.Bool("sync", false, "sync mode")
	flag.Parse()

	if START_PORT < 4000 || END_PORT > 4000+CGS-1 || CGS < 1 || START_PORT > END_PORT || MAX_ALLOWED_PEERS < 1 || MAX_ALLOWED_PEERS > CGS-1 {
		fmt.Println("Eroare parametrii de intrare")
		return
	}

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

	consensusGroup := make([]string, 0)

	for i := 4000; i < 4000+CGS; i++ {
		consensusGroup = append(consensusGroup, p2p.NewConnectParamsFromPort(i).ID.Pretty())
	}

	v := validators.Validators{}
	v.SetConsensusGroup(consensusGroup)

	var chrs []*chronology.Chronology

	for i := 0; i < len(nodes); i++ {
		v.SetSelf(consensusGroup[START_PORT-4000+i])
		chr := chronology.New(nodes[i], &v, GENESIS_TIME_STAMP, ROUND_DURATION)
		chr.DoLog = i == 0
		chr.DoSyncMode = SYNC
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
