package spos_test

import (
	"context"
	"crypto"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
)

const (
	Memory int = iota
	Network
)

func InitMessage() []*spos.Message {
	consensusGroupSize := 9
	firstNodeId := 1
	lastNodeId := 9
	maxAllowedPeers := 3
	roundDuration := 100 * time.Millisecond

	PBFTThreshold := consensusGroupSize*2/3 + 1

	genesisTime := time.Now()
	currentTime := genesisTime

	// start P2P
	nodes := startP2PConnections(firstNodeId, lastNodeId, maxAllowedPeers)

	// create consensus group list
	consensusGroup := createConsensusGroup(consensusGroupSize)

	// create instances
	var msgs []*spos.Message

	for i := 0; i < len(nodes); i++ {
		log := i == 0

		rnd := chronology.NewRound(
			genesisTime,
			currentTime,
			roundDuration)

		syncTime := &ntp.LocalTime{}
		syncTime.SetClockOffset(0)

		chr := chronology.NewChronology(
			log,
			true,
			rnd,
			genesisTime,
			syncTime)

		vld := spos.NewValidators(
			nil,
			nil,
			consensusGroup,
			consensusGroup[firstNodeId+i-1])

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

		dta := []byte("X")

		cns := spos.NewConsensus(
			log,
			&dta,
			vld,
			rth,
			rnds,
			chr)

		msg := spos.NewMessage(nodes[i], cns)

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrStartRound),
			chronology.SubroundId(spos.SrBlock),
			int64(roundDuration*5/100),
			cns.GetSubroundName(spos.SrStartRound),
			msg.StartRound,
			nil,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrBlock),
			chronology.SubroundId(spos.SrCommitmentHash),
			int64(roundDuration*25/100),
			cns.GetSubroundName(spos.SrBlock),
			msg.SendBlock, msg.ExtendBlock,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrCommitmentHash),
			chronology.SubroundId(spos.SrBitmap),
			int64(roundDuration*40/100),
			cns.GetSubroundName(spos.SrCommitmentHash),
			msg.SendCommitmentHash,
			msg.ExtendCommitmentHash,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrBitmap),
			chronology.SubroundId(spos.SrCommitment),
			int64(roundDuration*55/100),
			cns.GetSubroundName(spos.SrBitmap),
			msg.SendBitmap,
			msg.ExtendBitmap,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrCommitment),
			chronology.SubroundId(spos.SrSignature),
			int64(roundDuration*70/100),
			cns.GetSubroundName(spos.SrCommitment),
			msg.SendCommitment,
			msg.ExtendCommitment,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrSignature),
			chronology.SubroundId(spos.SrEndRound),
			int64(roundDuration*85/100),
			cns.GetSubroundName(spos.SrSignature),
			msg.SendSignature,
			msg.ExtendSignature,
			cns.CheckConsensus))

		chr.AddSubround(spos.NewSubround(
			chronology.SubroundId(spos.SrEndRound),
			-1,
			int64(roundDuration*100/100),
			cns.GetSubroundName(spos.SrEndRound),
			msg.EndRound,
			msg.ExtendEndRound,
			cns.CheckConsensus))

		msgs = append(msgs, msg)
	}

	// start Chr for each node
	//for i := 0; i < len(nodes); i++ {
	//	go msgs[i].Cns.Chr.StartRounds()
	//}

	return msgs
}

func startP2PConnections(firstNodeId int, lastNodeId int, maxAllowedPeers int) []*p2p.Messenger {
	marsh := &mock.MarshalizerMock{}

	var nodes []*p2p.Messenger

	for i := firstNodeId; i <= lastNodeId; i++ {
		node, err := createMessenger(Memory, 4000+i-1, maxAllowedPeers, marsh)

		if err != nil {
			continue
		}

		nodes = append(nodes, &node)
	}

	//time.Sleep(time.Second)

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

	return nodes
}

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

func createConsensusGroup(consensusGroupSize int) []string {
	consensusGroup := make([]string, 0)

	for i := 1; i <= consensusGroupSize; i++ {
		consensusGroup = append(consensusGroup, p2p.NewConnectParamsFromPort(4000+i-1).ID.Pretty())
	}

	return consensusGroup
}

func TestNewMessage(t *testing.T) {
	consensusGroup := []string{"1", "2", "3"}

	vld := spos.NewValidators(
		nil,
		nil,
		consensusGroup,
		consensusGroup[0])

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetAgreement(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetAgreement(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetAgreement(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetAgreement(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetAgreement(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	cns := spos.NewConsensus(
		true,
		nil,
		vld,
		nil,
		nil,
		nil)

	msg := spos.NewMessage(nil, nil)

	assert.Equal(t, 0, cap(msg.ChRcvMsg[spos.MtBlockHeader]))

	msg2 := spos.NewMessage(nil, cns)

	assert.Equal(t, len(cns.Validators.ConsensusGroup()), cap(msg2.ChRcvMsg[spos.MtBlockHeader]))
}

func TestMessage_StartRound(t *testing.T) {
	msgs := InitMessage()

	r := msgs[0].StartRound()
	assert.Equal(t, true, r)
}

func TestMessage_EndRound(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Hdr = &block.Header{}

	r := msgs[0].EndRound()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r = msgs[0].EndRound()
	assert.Equal(t, true, r)

	msgs[0].Cns.Validators.SetSelf(msgs[0].Cns.Validators.ConsensusGroup()[1])

	r = msgs[0].EndRound()
	assert.Equal(t, 2, msgs[0].RoundsWithBlock)
	assert.Equal(t, true, r)
}

func TestMessage_SendBlock(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)

	r := msgs[0].SendBlock()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrBlock, spos.SsNotFinished)
	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrBlock, true)

	r = msgs[0].SendBlock()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrBlock, false)
	msgs[0].Cns.Validators.SetSelf(msgs[0].Cns.Validators.ConsensusGroup()[1])

	r = msgs[0].SendBlock()
	assert.Equal(t, false, r)

	msgs[0].Cns.Validators.SetSelf(msgs[0].Cns.Validators.ConsensusGroup()[0])

	r = msgs[0].SendBlock()
	assert.Equal(t, true, r)
	assert.Equal(t, uint64(1), msgs[0].Hdr.Nonce)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrBlock, false)
	msgs[0].Blkc.CurrentBlock = msgs[0].Hdr

	r = msgs[0].SendBlock()
	assert.Equal(t, true, r)
	assert.Equal(t, uint64(2), msgs[0].Hdr.Nonce)
}

func TestMessage_SendCommitmentHash(t *testing.T) {
	msgs := InitMessage()

	r := msgs[0].SendCommitmentHash()
	assert.Equal(t, true, r)

	msgs[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)

	r = msgs[0].SendCommitmentHash()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)
	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrCommitmentHash, true)

	r = msgs[0].SendCommitmentHash()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrCommitmentHash, false)
	msgs[0].Cns.Data = nil

	r = msgs[0].SendCommitmentHash()
	assert.Equal(t, false, r)

	dta := []byte("X")
	msgs[0].Cns.Data = &dta

	r = msgs[0].SendCommitmentHash()
	assert.Equal(t, true, r)
}

func TestMessage_SendBitmap(t *testing.T) {
	msgs := InitMessage()

	r := msgs[0].SendBitmap()
	assert.Equal(t, true, r)

	msgs[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)

	r = msgs[0].SendBitmap()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)
	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrBitmap, true)

	r = msgs[0].SendBitmap()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrBitmap, false)
	msgs[0].Cns.Validators.SetSelf(msgs[0].Cns.Validators.ConsensusGroup()[1])

	r = msgs[0].SendBitmap()
	assert.Equal(t, false, r)

	msgs[0].Cns.Validators.SetSelf(msgs[0].Cns.Validators.ConsensusGroup()[0])
	msgs[0].Cns.Data = nil

	r = msgs[0].SendBitmap()
	assert.Equal(t, false, r)

	dta := []byte("X")
	msgs[0].Cns.Data = &dta
	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrCommitmentHash, true)

	r = msgs[0].SendBitmap()
	assert.Equal(t, true, r)
	assert.Equal(t, true, msgs[0].Cns.Agreement(msgs[0].Cns.Self(), spos.SrBitmap))
}

func TestMessage_SendCommitment(t *testing.T) {
	msgs := InitMessage()

	r := msgs[0].SendCommitment()
	assert.Equal(t, true, r)

	msgs[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)

	r = msgs[0].SendCommitment()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)
	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrCommitment, true)

	r = msgs[0].SendCommitment()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrCommitment, false)

	r = msgs[0].SendCommitment()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrBitmap, true)
	msgs[0].Cns.Data = nil

	r = msgs[0].SendCommitment()
	assert.Equal(t, false, r)

	dta := []byte("X")
	msgs[0].Cns.Data = &dta

	r = msgs[0].SendCommitment()
	assert.Equal(t, true, r)
}

func TestMessage_SendSignature(t *testing.T) {
	msgs := InitMessage()

	r := msgs[0].SendSignature()
	assert.Equal(t, true, r)

	msgs[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	msgs[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r = msgs[0].SendSignature()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrSignature, spos.SsNotFinished)
	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrSignature, true)

	r = msgs[0].SendSignature()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrSignature, false)

	r = msgs[0].SendSignature()
	assert.Equal(t, false, r)

	msgs[0].Cns.SetAgreement(msgs[0].Cns.Self(), spos.SrBitmap, true)
	msgs[0].Cns.Data = nil

	r = msgs[0].SendSignature()
	assert.Equal(t, false, r)

	dta := []byte("X")
	msgs[0].Cns.Data = &dta

	r = msgs[0].SendSignature()
	assert.Equal(t, true, r)
}

func TestMessage_BroadcastMessage(t *testing.T) {
	msgs := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err := marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtBlockHeader,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	r := msgs[0].BroadcastMessage(cnsDta)
	assert.Equal(t, true, r)
}

func TestMessage_ExtendBlock(t *testing.T) {
	msgs := InitMessage()

	msgs[0].ExtendBlock()
	assert.Equal(t, spos.SsExtended, msgs[0].Cns.Status(spos.SrBlock))
}

func TestMessage_ExtendCommitmentHash(t *testing.T) {
	msgs := InitMessage()

	msgs[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, msgs[0].Cns.Status(spos.SrCommitmentHash))

	for i := 0; i < len(msgs[0].Cns.Validators.ConsensusGroup()); i++ {
		msgs[0].Cns.SetAgreement(msgs[0].Cns.Validators.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	msgs[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, msgs[0].Cns.Status(spos.SrCommitmentHash))
}

func TestMessage_ExtendBitmap(t *testing.T) {
	msgs := InitMessage()

	msgs[0].ExtendBitmap()
	assert.Equal(t, spos.SsExtended, msgs[0].Cns.Status(spos.SrBitmap))
}

func TestMessage_ExtendCommitment(t *testing.T) {
	msgs := InitMessage()

	msgs[0].ExtendCommitment()
	assert.Equal(t, spos.SsExtended, msgs[0].Cns.Status(spos.SrCommitment))
}

func TestMessage_ExtendSignature(t *testing.T) {
	msgs := InitMessage()

	msgs[0].ExtendSignature()
	assert.Equal(t, spos.SsExtended, msgs[0].Cns.Status(spos.SrSignature))
}

func TestMessage_ExtendEndRound(t *testing.T) {
	msgs := InitMessage()

	msgs[0].ExtendEndRound()
}

func TestMessage_ReceiveMessage(t *testing.T) {
	msgs := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err := marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtBlockHeader,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m := p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	hops := m.Hops
	msgs[0].ReceiveMessage(*msgs[0].P2p, (*msgs[0].P2p).ID().Pretty(), m)
	assert.Equal(t, hops+1, m.Hops)
}

func TestMessage_ConsumeReceivedMessage(t *testing.T) {
	msgs := InitMessage()

	// Received BLOCK
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err := marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtBlockHeader,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m := p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	msgs[0].ConsumeReceivedMessage(&m.Payload)

	// Received COMMITMENT_HASH
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtCommitmentHash,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m = p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	msgs[0].ConsumeReceivedMessage(&m.Payload)

	// Received BITMAP
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtBitmap,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m = p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	msgs[0].ConsumeReceivedMessage(&m.Payload)

	// Received COMMITMENT
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtCommitment,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m = p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	msgs[0].ConsumeReceivedMessage(&m.Payload)

	// Received SIGNATURE
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtSignature,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m = p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	msgs[0].ConsumeReceivedMessage(&m.Payload)

	// Received UNKNOWN
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtUnknown,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m = p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	msgs[0].ConsumeReceivedMessage(&m.Payload)
}

func TestMessage_DecodeMessage(t *testing.T) {
	msgs := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err := marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtBlockHeader,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	message, err = marshal.DefMarsh.Marshal(cnsDta)

	assert.Nil(t, err)

	m := p2p.NewMessage((*msgs[0].P2p).ID().Pretty(), message, marshal.DefMarsh)

	msgType, msgData := msgs[0].DecodeMessage(nil)

	assert.Equal(t, spos.MtUnknown, msgType)
	assert.Nil(t, msgData)

	msgType, msgData = msgs[0].DecodeMessage(&m.Payload)

	assert.Equal(t, spos.MtBlockHeader, msgType)
	assert.Equal(t, []byte(msgs[0].Cns.Self()), msgData.Signature)
}

func TestMessage_GetMessageTypeName(t *testing.T) {
	msgs := InitMessage()

	r := msgs[0].GetMessageTypeName(spos.MtBlockBody)
	assert.Equal(t, "<BLOCK_BODY>", r)

	r = msgs[0].GetMessageTypeName(spos.MtBlockHeader)
	assert.Equal(t, "<BLOCK_HEADER>", r)

	r = msgs[0].GetMessageTypeName(spos.MtCommitmentHash)
	assert.Equal(t, "<COMMITMENT_HASH>", r)

	r = msgs[0].GetMessageTypeName(spos.MtBitmap)
	assert.Equal(t, "<BITMAP>", r)

	r = msgs[0].GetMessageTypeName(spos.MtCommitment)
	assert.Equal(t, "<COMMITMENT>", r)

	r = msgs[0].GetMessageTypeName(spos.MtSignature)
	assert.Equal(t, "<SIGNATURE>", r)

	r = msgs[0].GetMessageTypeName(spos.MtUnknown)
	assert.Equal(t, "<UNKNOWN>", r)

	r = msgs[0].GetMessageTypeName(spos.MessageType(-1))
	assert.Equal(t, "Undifined message type", r)
}

func TestMessage_CheckChannels(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(msgs[0].Cns.Chr.Round().TimeDuration()))

	assert.Equal(t, false, msgs[0].Cns.ShouldCheckConsensus())

	// BLOCK BODY
	blk := &block.Block{}

	message, err := marshal.DefMarsh.Marshal(blk)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.Self()),
		spos.MtBlockBody,
		[]byte(msgs[0].Cns.Chr.SyncTime().CurrentTime(msgs[0].Cns.Chr.ClockOffset()).String()))

	msgs[0].ChRcvMsg[spos.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, msgs[0].Cns.ShouldCheckConsensus())
	assert.Equal(t, false, msgs[0].Cns.Validators.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrBlock))

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().CurrentTime(msgs[0].Cns.Chr.ClockOffset()).String())

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtBlockHeader,
		[]byte(msgs[0].Cns.Chr.SyncTime().CurrentTime(msgs[0].Cns.Chr.ClockOffset()).String()))

	msgs[0].ChRcvMsg[spos.MtBlockHeader] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, msgs[0].Cns.ShouldCheckConsensus())
	assert.Equal(t, true, msgs[0].Cns.Validators.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrBlock))

	msgs[0].Cns.SetShouldCheckConsensus(false)

	// COMMITMENT_HASH
	cnsDta = spos.NewConsensusData(
		*msgs[0].Cns.Data,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitmentHash,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].ChRcvMsg[spos.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, msgs[0].Cns.ShouldCheckConsensus())
	assert.Equal(t, true, msgs[0].Cns.Validators.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrCommitmentHash))

	msgs[0].Cns.SetShouldCheckConsensus(false)

	// BITMAP
	pks := make([][]byte, 0)

	for i := 0; i < len(msgs[0].Cns.ConsensusGroup()); i++ {
		pks = append(pks, []byte(msgs[0].Cns.ConsensusGroup()[i]))
	}

	cnsDta = spos.NewConsensusData(
		*msgs[0].Cns.Data,
		pks,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtBitmap,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].ChRcvMsg[spos.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, msgs[0].Cns.ShouldCheckConsensus())

	for i := 0; i < len(msgs[0].Cns.ConsensusGroup()); i++ {
		assert.Equal(t, true, msgs[0].Cns.Validators.Agreement(msgs[0].Cns.ConsensusGroup()[i], spos.SrBitmap))
	}

	msgs[0].Cns.SetShouldCheckConsensus(false)

	// COMMITMENT
	cnsDta = spos.NewConsensusData(
		*msgs[0].Cns.Data,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitment,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].ChRcvMsg[spos.MtCommitment] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, msgs[0].Cns.ShouldCheckConsensus())
	assert.Equal(t, true, msgs[0].Cns.Validators.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrCommitment))

	msgs[0].Cns.SetShouldCheckConsensus(false)

	// SIGNATURE
	cnsDta = spos.NewConsensusData(
		*msgs[0].Cns.Data,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtSignature,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].ChRcvMsg[spos.MtSignature] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, msgs[0].Cns.ShouldCheckConsensus())
	assert.Equal(t, true, msgs[0].Cns.Validators.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrSignature))
}

func TestMessage_ReceivedBlock(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(msgs[0].Cns.Chr.Round().TimeDuration()))

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	message, err := marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtBlockBody,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	r := msgs[0].ReceivedBlockBody(cnsDta)
	assert.Equal(t, true, r)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtBlockHeader,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)

	r = msgs[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrBlock, spos.SsNotFinished)

	hdr.PrevHash = []byte("X")
	message, err = marshal.DefMarsh.Marshal(hdr)
	assert.Nil(t, err)

	cnsDta.Data = message

	r = msgs[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, false, r)

	hdr.PrevHash = []byte("")
	message, err = marshal.DefMarsh.Marshal(hdr)
	assert.Nil(t, err)

	cnsDta.Data = message

	r = msgs[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, msgs[0].Cns.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrBlock))
}

func TestMessage_ReceivedCommitmentHash(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(msgs[0].Cns.Chr.Round().TimeDuration()))

	dta := []byte("X")

	cnsDta := spos.NewConsensusData(
		dta,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitmentHash,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	for i := 0; i < msgs[0].Cns.Threshold(spos.SrCommitmentHash); i++ {
		msgs[0].Cns.Validators.SetAgreement(msgs[0].Cns.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	r := msgs[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, false, r)

	for i := 0; i < msgs[0].Cns.Threshold(spos.SrCommitmentHash); i++ {
		msgs[0].Cns.Validators.SetAgreement(msgs[0].Cns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
	}

	msgs[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)

	r = msgs[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	r = msgs[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, msgs[0].Cns.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrCommitmentHash))
}

func TestMessage_ReceivedBitmap(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(msgs[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*msgs[0].Cns.Data,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitmentHash,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)

	r := msgs[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)

	pks := make([][]byte, 0)

	for i := 0; i < msgs[0].Cns.Threshold(spos.SrBitmap)-1; i++ {
		pks = append(pks, []byte(msgs[0].Cns.ConsensusGroup()[i]))
	}

	cnsDta.PubKeys = pks

	r = msgs[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)
	assert.Equal(t, chronology.SubroundId(-1), msgs[0].Cns.Chr.SelfSubround())

	cnsDta.PubKeys = append(cnsDta.PubKeys, []byte(msgs[0].Cns.ConsensusGroup()[msgs[0].Cns.Threshold(spos.SrBitmap)-1]))

	r = msgs[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, msgs[0].Cns.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrBitmap))

	for i := 0; i < msgs[0].Cns.Threshold(spos.SrBitmap); i++ {
		assert.Equal(t, true, msgs[0].Cns.Validators.Agreement(msgs[0].Cns.ConsensusGroup()[i], spos.SrBitmap))
	}

	cnsDta.PubKeys = append(cnsDta.PubKeys, []byte("X"))

	r = msgs[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)
	assert.Equal(t, chronology.SubroundId(-1), msgs[0].Cns.Chr.SelfSubround())
}

func TestMessage_ReceivedCommitment(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(msgs[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*msgs[0].Cns.Data,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitment,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)

	r := msgs[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)

	r = msgs[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, false, r)

	msgs[0].Cns.Validators.SetAgreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrBitmap, true)

	r = msgs[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, msgs[0].Cns.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrCommitment))
}

func TestMessage_ReceivedSignature(t *testing.T) {
	msgs := InitMessage()

	msgs[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(msgs[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*msgs[0].Cns.Data,
		nil,
		[]byte(msgs[0].Cns.ConsensusGroup()[1]),
		spos.MtSignature,
		[]byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset())))

	msgs[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r := msgs[0].ReceivedSignature(cnsDta)
	assert.Equal(t, false, r)

	msgs[0].Cns.SetStatus(spos.SrSignature, spos.SsNotFinished)

	r = msgs[0].ReceivedSignature(cnsDta)
	assert.Equal(t, false, r)

	msgs[0].Cns.Validators.SetAgreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrBitmap, true)

	r = msgs[0].ReceivedSignature(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, msgs[0].Cns.Agreement(msgs[0].Cns.ConsensusGroup()[1], spos.SrSignature))
}

func TestMessage_CheckIfBlockIsValid(t *testing.T) {
	msgs := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	hdr.PrevHash = []byte("X")

	r := msgs[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.PrevHash = []byte("")

	r = msgs[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 2

	r = msgs[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 1
	msgs[0].Blkc.CurrentBlock = hdr

	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(msgs[0].Cns.Chr.SyncTime().FormatedCurrentTime(msgs[0].Cns.Chr.ClockOffset()))

	r = msgs[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")

	r = msgs[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")

	r = msgs[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 2

	r = msgs[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)
}
