package integrationTests

import (
	"context"
	"crypto"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/statistic"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	Memory int = iota
	Network
)

func TestChronology(t *testing.T) {

	GENESIS_TIME := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
	CONSENSUS_GROUP_SIZE := 21
	MAX_ALLOWED_PEERS := 4
	ROUND_DURATION := time.Duration(4000 * time.Millisecond)
	FIRST_NODE_ID := 1
	LAST_NODE_ID := 21
	SYNC_MODE := true
	PBFTThreshold := CONSENSUS_GROUP_SIZE*2/3 + 1

	// start P2P
	nodes := startP2PConnections(FIRST_NODE_ID, LAST_NODE_ID, MAX_ALLOWED_PEERS)

	// create consensus group list
	consensusGroup := createConsensusGroup(CONSENSUS_GROUP_SIZE)

	// create Chronology (set ChronologyIn parameters) for each node
	var chrs []*chronology.Chronology

	for i := 0; i < len(nodes); i++ {

		// set ChronologyIn parameters
		syncTime := ntp.NewSyncTime(ROUND_DURATION, ntp.Query)
		block := block.New(-1, "", "", "", "", "")
		blockChain := blockchain.New(nil, syncTime, i == 0)
		validators := validators.New(consensusGroup, consensusGroup[FIRST_NODE_ID+i-1])
		consensus := consensus.New(consensus.Threshold{1, PBFTThreshold, PBFTThreshold, PBFTThreshold, PBFTThreshold})
		round := round.NewRoundFromDateTime(GENESIS_TIME, syncTime.CurrentTime(), ROUND_DURATION, createRoundTimeDivision(ROUND_DURATION), round.Subround{round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED})
		statistic := statistic.New()

		// create ChronologyIn
		chronologyIn := chronology.ChronologyIn{GenesisTime: GENESIS_TIME, P2PNode: nodes[i], Block: &block, BlockChain: &blockChain, Validators: &validators, Consensus: &consensus, Round: &round, Statistic: &statistic, SyncTime: syncTime}

		// create Chronology
		chr := chronology.New(&chronologyIn)

		chr.DoLog = i == 0
		chr.DoSyncMode = SYNC_MODE

		chrs = append(chrs, chr)
	}

	// start Chronology for each node
	for i := 0; i < len(nodes); i++ {
		go chrs[i].StartRounds()
	}

	// log blockChain for first node
	logBlockChain(chrs[0])

	// close P2P
	for i := 0; i < len(nodes); i++ {
		chrs[i].DoRun = false
		(*nodes[i]).Close()
	}
}

func startP2PConnections(FIRST_NODE_ID int, LAST_NODE_ID int, MAX_ALLOWED_PEERS int) []*p2p.Messenger {
	marsh := &mock.MockMarshalizer{}

	var nodes []*p2p.Messenger

	for i := FIRST_NODE_ID; i <= LAST_NODE_ID; i++ {
		node, err := createMessenger(Network, 4000+i-1, MAX_ALLOWED_PEERS, marsh)

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

	return nodes
}

func createConsensusGroup(CONSENSUS_GROUP_SIZE int) []string {
	consensusGroup := make([]string, 0)

	for i := 1; i <= CONSENSUS_GROUP_SIZE; i++ {
		consensusGroup = append(consensusGroup, p2p.NewConnectParamsFromPort(4000+i-1).ID.Pretty())
	}

	return consensusGroup
}

func logBlockChain(chr *chronology.Chronology) {
	oldNounce := -1

	for {
		time.Sleep(100 * time.Millisecond)

		currentBlock := chr.BlockChain.GetCurrentBlock()
		if currentBlock == nil {
			continue
		}

		if currentBlock.Nonce > oldNounce {
			oldNounce = currentBlock.Nonce
			spew.Dump(currentBlock)
			fmt.Printf("\n********** There was %d rounds and was proposed %d blocks, which means %.2f%% hit rate **********\n", chr.Statistic.GetRounds(), chr.Statistic.GetRoundsWithBlock(), float64(chr.Statistic.GetRoundsWithBlock())*100/float64(chr.Statistic.GetRounds()))
		}

		if oldNounce >= 2 {
			break
		}
	}
}

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

func createRoundTimeDivision(duration time.Duration) []time.Duration {

	var d []time.Duration

	for i := round.RS_START_ROUND; i <= round.RS_END_ROUND; i++ {
		switch i {
		case round.RS_START_ROUND:
			d = append(d, time.Duration(5*duration/100))
		case round.RS_BLOCK:
			d = append(d, time.Duration(25*duration/100))
		case round.RS_COMITMENT_HASH:
			d = append(d, time.Duration(40*duration/100))
		case round.RS_BITMAP:
			d = append(d, time.Duration(55*duration/100))
		case round.RS_COMITMENT:
			d = append(d, time.Duration(70*duration/100))
		case round.RS_SIGNATURE:
			d = append(d, time.Duration(85*duration/100))
		case round.RS_END_ROUND:
			d = append(d, time.Duration(100*duration/100))
		}
	}

	return d
}
