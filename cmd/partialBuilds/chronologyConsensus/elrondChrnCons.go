package main

import (
	"context"
	"crypto"
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	//beevik "github.com/beevik/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	// Memory is used to create a p2p in memory
	Memory int = iota
	// Network is used to make a p2p over the network
	Network
)

func main() {

	genesisTime := time.Date(time.Now().Year(),
		time.Now().Month(),
		time.Now().Day(),
		0,
		0,
		0,
		0,
		time.Local)

	// Parse options from the command line
	consensusGroupSize := flag.Int("size",
		21,
		"consensus group size which validate proposed block by the leader")

	maxAllowedPeers := flag.Int("peers",
		4,
		"max connections allowed by each peer")

	roundDuration := flag.Duration("duration",
		1000*time.Millisecond,
		"round duration in milliseconds")

	firstNodeId := flag.Int("first",
		1,
		"first node ID. This ID should be between 1 and consensus group size")

	lastNodeId := flag.Int("last",
		21,
		"last node ID. This ID should be between 1 and consensus group size, but also greater or equal than first node ID")

	syncMode := flag.Bool("sync",
		false,
		"sync mode in subrounds will be used")

	flag.Parse()

	if *firstNodeId < 1 ||
		*lastNodeId > *consensusGroupSize ||
		*consensusGroupSize < 1 ||
		*firstNodeId > *lastNodeId ||
		*maxAllowedPeers < 1 ||
		*maxAllowedPeers > *consensusGroupSize-1 {
		fmt.Println("Eroare parametrii de intrare")
		return
	} else {
		fmt.Printf("size = %d\npeers = %d\nduration = %d\nfirst = %d\nlast = %d\nsynctime = %v\n\n",
			*consensusGroupSize, *maxAllowedPeers, *roundDuration, *firstNodeId, *lastNodeId, *syncMode)
	}

	PBFTThreshold := *consensusGroupSize*2/3 + 1

	currentTime := time.Now()

	// start P2P
	nodes := startP2PConnections(*firstNodeId, *lastNodeId, *maxAllowedPeers)

	// create consensus group list
	consensusGroup := createConsensusGroup(*consensusGroupSize)

	// create instances
	var msgs []*spos.Message

	for i := 0; i < len(nodes); i++ {
		log := i == 0

		rnd := chronology.NewRound(
			genesisTime,
			currentTime,
			*roundDuration)

		//syncTime := ntp.NewSyncTime(*roundDuration, beevik.Query)
		syncTime := &ntp.LocalTime{}
		syncTime.SetClockOffset(0)

		chr := chronology.NewChronology(
			log,
			*syncMode,
			rnd,
			genesisTime,
			syncTime)

		vld := spos.NewValidators(
			nil,
			nil,
			consensusGroup,
			consensusGroup[*firstNodeId+i-1])

		for j := 0; j < len(vld.ConsensusGroup()); j++ {
			vld.SetAgreement(vld.ConsensusGroup()[j], spos.SrBlock, false)
			vld.SetAgreement(vld.ConsensusGroup()[j], spos.SrCommitmentHash, false)
			vld.SetAgreement(vld.ConsensusGroup()[j], spos.SrBitmap, false)
			vld.SetAgreement(vld.ConsensusGroup()[j], spos.SrCommitment, false)
			vld.SetAgreement(vld.ConsensusGroup()[j], spos.SrSignature, false)
		}

		rth := spos.NewRoundThreshold()

		rth.SetThreshold(spos.SrBlock, 1)
		rth.SetThreshold(spos.SrCommitmentHash, PBFTThreshold)
		rth.SetThreshold(spos.SrBitmap, PBFTThreshold)
		rth.SetThreshold(spos.SrCommitment, PBFTThreshold)
		rth.SetThreshold(spos.SrSignature, PBFTThreshold)

		rnds := spos.NewRoundStatus()

		rnds.SetStatus(spos.SrBlock, spos.SsNotFinished)
		rnds.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)
		rnds.SetStatus(spos.SrBitmap, spos.SsNotFinished)
		rnds.SetStatus(spos.SrCommitment, spos.SsNotFinished)
		rnds.SetStatus(spos.SrSignature, spos.SsNotFinished)

		cns := spos.NewConsensus(
			log,
			nil,
			vld,
			rth,
			rnds,
			chr)

		msg := spos.NewMessage(nodes[i], cns)

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrStartRound),
			chronology.SubroundId(spos.SrBlock),
			int64(*roundDuration*5/100),
			cns.GetSubroundName(spos.SrStartRound),
			msg.StartRound,
			nil,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrBlock),
			chronology.SubroundId(spos.SrCommitmentHash),
			int64(*roundDuration*25/100),
			cns.GetSubroundName(spos.SrBlock),
			msg.SendBlock,
			msg.ExtendBlock,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrCommitmentHash),
			chronology.SubroundId(spos.SrBitmap),
			int64(*roundDuration*40/100),
			cns.GetSubroundName(spos.SrCommitmentHash),
			msg.SendCommitmentHash,
			msg.ExtendCommitmentHash,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrBitmap),
			chronology.SubroundId(spos.SrCommitment),
			int64(*roundDuration*55/100),
			cns.GetSubroundName(spos.SrBitmap),
			msg.SendBitmap,
			msg.ExtendBitmap,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrCommitment),
			chronology.SubroundId(spos.SrSignature),
			int64(*roundDuration*70/100),
			cns.GetSubroundName(spos.SrCommitment),
			msg.SendCommitment,
			msg.ExtendCommitment,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrSignature),
			chronology.SubroundId(spos.SrEndRound),
			int64(*roundDuration*85/100),
			cns.GetSubroundName(spos.SrSignature),
			msg.SendSignature,
			msg.ExtendSignature,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrEndRound),
			-1,
			int64(*roundDuration*100/100),
			cns.GetSubroundName(spos.SrEndRound),
			msg.EndRound,
			msg.ExtendEndRound,
			cns.CheckConsensus))

		msgs = append(msgs, msg)
	}

	// start Chr for each node
	for i := 0; i < len(nodes); i++ {
		go msgs[i].Cns.Chr.StartRounds()
	}

	// log BlockChain for first node
	logBlockChain(msgs[0])

	// close P2P
	for i := 0; i < len(nodes); i++ {
		msgs[i].Cns.Chr.DoRun = false
		(*nodes[i]).Close()
	}
}

func startP2PConnections(firstNodeId int, lastNodeId int, maxAllowedPeers int) []*p2p.Messenger {
	marsh := &mock.MarshalizerMock{}

	var nodes []*p2p.Messenger

	for i := firstNodeId; i <= lastNodeId; i++ {
		node, err := createMessenger(Network, 4000+i-1, maxAllowedPeers, marsh)
		//node, err := createMessenger(Memory, 4000+i-1, maxAllowedPeers, marsh)

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

func createConsensusGroup(consensusGroupSize int) []string {
	consensusGroup := make([]string, 0)

	for i := 1; i <= consensusGroupSize; i++ {
		consensusGroup = append(consensusGroup, p2p.NewConnectParamsFromPort(4000+i-1).ID.Pretty())
	}

	return consensusGroup
}

func logBlockChain(msg *spos.Message) {
	oldNounce := uint64(0)

	for {
		time.Sleep(100 * time.Millisecond)

		currentBlock := msg.Blkc.CurrentBlock
		if currentBlock == nil {
			continue
		}

		if currentBlock.Nonce > oldNounce {
			oldNounce = currentBlock.Nonce
			spew.Dump(currentBlock)
			fmt.Printf("\n********** There was %d rounds and was proposed %d blocks, which means %.2f%% hit rate **********\n",
				msg.Rounds, msg.RoundsWithBlock, float64(msg.RoundsWithBlock)*100/float64(msg.Rounds))
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
