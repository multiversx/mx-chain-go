package spos_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"testing"
	"time"

	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/stretchr/testify/assert"
)

func SendMessage(msg []byte) {
	fmt.Println(msg)
}

func InitMessage() []*spos.SPOSConsensusWorker {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond

	PBFTThreshold := consensusGroupSize*2/3 + 1

	genesisTime := time.Now()
	currentTime := genesisTime

	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	// create instances
	var coms []*spos.SPOSConsensusWorker

	for i := 0; i < consensusGroupSize; i++ {
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

		vld := spos.NewRoundConsensus(
			consensusGroup,
			consensusGroup[i])

		for j := 0; j < len(vld.ConsensusGroup()); j++ {
			vld.SetJobDone(vld.ConsensusGroup()[j], spos.SrBlock, false)
			vld.SetJobDone(vld.ConsensusGroup()[j], spos.SrCommitmentHash, false)
			vld.SetJobDone(vld.ConsensusGroup()[j], spos.SrBitmap, false)
			vld.SetJobDone(vld.ConsensusGroup()[j], spos.SrCommitment, false)
			vld.SetJobDone(vld.ConsensusGroup()[j], spos.SrSignature, false)
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

		blkc := blockchain.BlockChain{}
		txp := transactionPool.NewTransactionPool(nil)

		com := spos.NewCommunication(
			true,
			cns,
			&blkc,
			txp,
			mock.HasherMock{},
			mock.MarshalizerMock{})

		com.OnSendMessage = SendMessage

		GenerateSubRoundHandlers(roundDuration, cns, com)

		coms = append(coms, com)
	}

	return coms
}

// RoundTimeDuration defines the time duration in milliseconds of each round
const RoundTimeDuration = time.Duration(4000 * time.Millisecond)

func DoSubroundJob() bool {
	fmt.Printf("do job\n")
	time.Sleep(5 * time.Millisecond)
	return true
}

func DoExtendSubround() {
	fmt.Printf("do extend subround\n")
}

func DoCheckConsensusWithSuccess() bool {
	fmt.Printf("do check consensus with success in subround \n")
	return true
}

func DoCheckConsensusWithoutSuccess() bool {
	fmt.Printf("do check consensus without success in subround \n")
	return false
}

func GenerateSubRoundHandlers(roundDuration time.Duration, cns *spos.Consensus, com *spos.SPOSConsensusWorker) {
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrStartRound),
		chronology.SubroundId(spos.SrBlock),
		int64(roundDuration*5/100),
		cns.GetSubroundName(spos.SrStartRound),
		com.DoStartRoundJob,
		nil,
		func() bool { return true }))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBlock),
		chronology.SubroundId(spos.SrCommitmentHash),
		int64(roundDuration*25/100),
		cns.GetSubroundName(spos.SrBlock),
		com.DoBlockJob, com.ExtendBlock,
		cns.CheckBlockConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitmentHash),
		chronology.SubroundId(spos.SrBitmap),
		int64(roundDuration*40/100),
		cns.GetSubroundName(spos.SrCommitmentHash),
		com.DoCommitmentHashJob,
		com.ExtendCommitmentHash,
		cns.CheckCommitmentHashConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBitmap),
		chronology.SubroundId(spos.SrCommitment),
		int64(roundDuration*55/100),
		cns.GetSubroundName(spos.SrBitmap),
		com.DoBitmapJob,
		com.ExtendBitmap,
		cns.CheckBitmapConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitment),
		chronology.SubroundId(spos.SrSignature),
		int64(roundDuration*70/100),
		cns.GetSubroundName(spos.SrCommitment),
		com.DoCommitmentJob,
		com.ExtendCommitment,
		cns.CheckCommitmentConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrSignature),
		chronology.SubroundId(spos.SrEndRound),
		int64(roundDuration*85/100),
		cns.GetSubroundName(spos.SrSignature),
		com.DoSignatureJob,
		com.ExtendSignature,
		cns.CheckSignatureConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrEndRound),
		-1,
		int64(roundDuration*100/100),
		cns.GetSubroundName(spos.SrEndRound),
		com.DoEndRoundJob,
		com.ExtendEndRound,
		cns.CheckEndRoundConsensus))
}

func CreateConsensusGroup(consensusGroupSize int) []string {
	consensusGroup := make([]string, 0)

	for i := 0; i < consensusGroupSize; i++ {
		consensusGroup = append(consensusGroup, string(i+65))
	}

	return consensusGroup
}

func TestNewConsensusData(t *testing.T) {
	cnsData := spos.NewConsensusData(
		nil,
		nil,
		nil,
		spos.MtUnknown,
		nil)

	assert.NotNil(t, cnsData)
}

func TestNewMessage(t *testing.T) {
	consensusGroup := []string{"1", "2", "3"}

	vld := spos.NewRoundConsensus(
		consensusGroup,
		consensusGroup[0])

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	cns := spos.NewConsensus(
		true,
		nil,
		vld,
		nil,
		nil,
		nil)

	com := spos.NewCommunication(
		true,
		nil,
		nil,
		nil,
		mock.HasherMock{},
		mock.MarshalizerMock{})

	assert.Equal(t, 0, cap(com.ChRcvMsg[spos.MtBlockHeader]))

	msg2 := spos.NewCommunication(
		true,
		cns,
		nil,
		nil,
		mock.HasherMock{},
		mock.MarshalizerMock{})

	assert.Equal(t, len(cns.RoundConsensus.ConsensusGroup()), cap(msg2.ChRcvMsg[spos.MtBlockHeader]))
}

func TestMessage_StartRound(t *testing.T) {
	coms := InitMessage()

	r := coms[0].DoStartRoundJob()
	assert.Equal(t, true, r)
}

func TestMessage_EndRound(t *testing.T) {
	coms := InitMessage()

	coms[0].Hdr = &block.Header{}

	r := coms[0].DoEndRoundJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r = coms[0].DoEndRoundJob()
	assert.Equal(t, true, r)

	coms[0].Cns.RoundConsensus.SetSelfId(coms[0].Cns.RoundConsensus.ConsensusGroup()[1])

	r = coms[0].DoEndRoundJob()
	assert.Equal(t, 2, coms[0].RoundsWithBlock)
	assert.Equal(t, true, r)
}

func TestMessage_SendBlock(t *testing.T) {
	coms := InitMessage()

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(coms[0].Cns.Chr.Round().TimeDuration()))

	r := coms[0].DoBlockJob()
	assert.Equal(t, false, r)

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now())
	coms[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)

	r = coms[0].DoBlockJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrBlock, spos.SsNotFinished)
	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrBlock, true)

	r = coms[0].DoBlockJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrBlock, false)
	coms[0].Cns.RoundConsensus.SetSelfId(coms[0].Cns.RoundConsensus.ConsensusGroup()[1])

	r = coms[0].DoBlockJob()
	assert.Equal(t, false, r)

	coms[0].Cns.RoundConsensus.SetSelfId(coms[0].Cns.RoundConsensus.ConsensusGroup()[0])

	r = coms[0].DoBlockJob()
	assert.Equal(t, true, r)
	assert.Equal(t, uint64(1), coms[0].Hdr.Nonce)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrBlock, false)
	coms[0].Blkc.CurrentBlockHeader = coms[0].Hdr

	r = coms[0].DoBlockJob()
	assert.Equal(t, true, r)
	assert.Equal(t, uint64(2), coms[0].Hdr.Nonce)
}

func TestCommunication_CreateMiniBlocks(t *testing.T) {
	coms := InitMessage()
	coms[0].TxP.AddTransaction([]byte("tx_hash"), &transaction.Transaction{Nonce: 0}, 0)
	mblks := coms[0].CreateMiniBlocks()

	assert.Equal(t, 1, len(mblks))
	assert.Equal(t, []byte("tx_hash"), mblks[0].TxHashes[0])
}

func TestMessage_SendCommitmentHash(t *testing.T) {
	coms := InitMessage()

	r := coms[0].DoCommitmentHashJob()
	assert.Equal(t, true, r)

	coms[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)

	r = coms[0].DoCommitmentHashJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)
	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrCommitmentHash, true)

	r = coms[0].DoCommitmentHashJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrCommitmentHash, false)
	coms[0].Cns.Data = nil

	r = coms[0].DoCommitmentHashJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	coms[0].Cns.Data = &dta

	r = coms[0].DoCommitmentHashJob()
	assert.Equal(t, true, r)
}

func TestMessage_SendBitmap(t *testing.T) {
	coms := InitMessage()

	r := coms[0].DoBitmapJob()
	assert.Equal(t, true, r)

	coms[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)

	r = coms[0].DoBitmapJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)
	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrBitmap, true)

	r = coms[0].DoBitmapJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrBitmap, false)
	coms[0].Cns.RoundConsensus.SetSelfId(coms[0].Cns.RoundConsensus.ConsensusGroup()[1])

	r = coms[0].DoBitmapJob()
	assert.Equal(t, false, r)

	coms[0].Cns.RoundConsensus.SetSelfId(coms[0].Cns.RoundConsensus.ConsensusGroup()[0])
	coms[0].Cns.Data = nil

	r = coms[0].DoBitmapJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	coms[0].Cns.Data = &dta
	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrCommitmentHash, true)

	r = coms[0].DoBitmapJob()
	assert.Equal(t, true, r)
	assert.Equal(t, true, coms[0].Cns.GetJobDone(coms[0].Cns.SelfId(), spos.SrBitmap))
}

func TestMessage_SendCommitment(t *testing.T) {
	coms := InitMessage()

	r := coms[0].DoCommitmentJob()
	assert.Equal(t, true, r)

	coms[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)

	r = coms[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)
	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrCommitment, true)

	r = coms[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrCommitment, false)

	r = coms[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrBitmap, true)
	coms[0].Cns.Data = nil

	r = coms[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	coms[0].Cns.Data = &dta

	r = coms[0].DoCommitmentJob()
	assert.Equal(t, true, r)
}

func TestMessage_SendSignature(t *testing.T) {
	coms := InitMessage()

	r := coms[0].DoSignatureJob()
	assert.Equal(t, true, r)

	coms[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	coms[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r = coms[0].DoSignatureJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrSignature, spos.SsNotFinished)
	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrSignature, true)

	r = coms[0].DoSignatureJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrSignature, false)

	r = coms[0].DoSignatureJob()
	assert.Equal(t, false, r)

	coms[0].Cns.SetJobDone(coms[0].Cns.SelfId(), spos.SrBitmap, true)
	coms[0].Cns.Data = nil

	r = coms[0].DoSignatureJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	coms[0].Cns.Data = &dta

	r = coms[0].DoSignatureJob()
	assert.Equal(t, true, r)
}

func TestMessage_BroadcastMessage(t *testing.T) {
	coms := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtBlockHeader,
		[]byte(coms[0].GetTime()))

	coms[0].OnSendMessage = nil
	r := coms[0].BroadcastMessage(cnsDta)
	assert.Equal(t, false, r)

	coms[0].OnSendMessage = SendMessage
	r = coms[0].BroadcastMessage(cnsDta)
	assert.Equal(t, true, r)
}

func TestMessage_ExtendBlock(t *testing.T) {
	coms := InitMessage()

	coms[0].ExtendBlock()
	assert.Equal(t, spos.SsExtended, coms[0].Cns.Status(spos.SrBlock))
}

func TestMessage_ExtendCommitmentHash(t *testing.T) {
	coms := InitMessage()

	coms[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, coms[0].Cns.Status(spos.SrCommitmentHash))

	for i := 0; i < len(coms[0].Cns.RoundConsensus.ConsensusGroup()); i++ {
		coms[0].Cns.SetJobDone(coms[0].Cns.RoundConsensus.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	coms[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, coms[0].Cns.Status(spos.SrCommitmentHash))
}

func TestMessage_ExtendBitmap(t *testing.T) {
	coms := InitMessage()

	coms[0].ExtendBitmap()
	assert.Equal(t, spos.SsExtended, coms[0].Cns.Status(spos.SrBitmap))
}

func TestMessage_ExtendCommitment(t *testing.T) {
	coms := InitMessage()

	coms[0].ExtendCommitment()
	assert.Equal(t, spos.SsExtended, coms[0].Cns.Status(spos.SrCommitment))
}

func TestMessage_ExtendSignature(t *testing.T) {
	coms := InitMessage()

	coms[0].ExtendSignature()
	assert.Equal(t, spos.SsExtended, coms[0].Cns.Status(spos.SrSignature))
}

func TestMessage_ExtendEndRound(t *testing.T) {
	coms := InitMessage()

	coms[0].ExtendEndRound()
}

func TestMessage_ReceivedMessage(t *testing.T) {
	coms := InitMessage()

	// Received BLOCK_BODY
	blk := &block.TxBlockBody{}

	message, err := mock.MarshalizerMock{}.Marshal(blk)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtBlockBody,
		[]byte(coms[0].Cns.Chr.SyncTime().CurrentTime(coms[0].Cns.Chr.ClockOffset()).String()))

	coms[0].ReceivedMessage(cnsDta)

	// Received BLOCK_HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtBlockHeader,
		[]byte(coms[0].GetTime()))

	coms[0].ReceivedMessage(cnsDta)

	// Received COMMITMENT_HASH
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtCommitmentHash,
		[]byte(coms[0].GetTime()))

	coms[0].ReceivedMessage(cnsDta)

	// Received BITMAP
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtBitmap,
		[]byte(coms[0].GetTime()))

	coms[0].ReceivedMessage(cnsDta)

	// Received COMMITMENT
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtCommitment,
		[]byte(coms[0].GetTime()))

	coms[0].ReceivedMessage(cnsDta)

	// Received SIGNATURE
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtSignature,
		[]byte(coms[0].GetTime()))

	coms[0].ReceivedMessage(cnsDta)

	// Received UNKNOWN
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.SelfId()),
		spos.MtUnknown,
		[]byte(coms[0].GetTime()))

	coms[0].ReceivedMessage(cnsDta)
}

func TestMessage_DecodeBlockBody(t *testing.T) {
	coms := InitMessage()

	blk := &block.TxBlockBody{}

	mblks := make([]block.MiniBlock, 0)
	mblks = append(mblks, block.MiniBlock{DestShardID: 69})
	blk.MiniBlocks = mblks

	message, err := mock.MarshalizerMock{}.Marshal(blk)

	assert.Nil(t, err)

	dcdBlk := coms[0].DecodeBlockBody(nil)

	assert.Nil(t, dcdBlk)

	dcdBlk = coms[0].DecodeBlockBody(&message)

	assert.Equal(t, blk, dcdBlk)
	assert.Equal(t, uint32(69), dcdBlk.MiniBlocks[0].DestShardID)
}

func TestMessage_DecodeBlockHeader(t *testing.T) {
	coms := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())
	hdr.Signature = []byte(coms[0].Cns.SelfId())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	dcdHdr := coms[0].DecodeBlockHeader(nil)

	assert.Nil(t, dcdHdr)

	dcdHdr = coms[0].DecodeBlockHeader(&message)

	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte(coms[0].Cns.SelfId()), dcdHdr.Signature)
}

func TestMessage_CheckChannels(t *testing.T) {
	coms := InitMessage()

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(coms[0].Cns.Chr.Round().TimeDuration()))

	// BLOCK BODY
	blk := &block.TxBlockBody{}

	message, err := mock.MarshalizerMock{}.Marshal(blk)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtBlockBody,
		[]byte(coms[0].Cns.Chr.SyncTime().CurrentTime(coms[0].Cns.Chr.ClockOffset()).String()))

	coms[0].ChRcvMsg[spos.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, coms[0].Cns.RoundConsensus.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrBlock))

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].Cns.Chr.SyncTime().CurrentTime(coms[0].Cns.Chr.ClockOffset()).String())

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtBlockHeader,
		[]byte(coms[0].Cns.Chr.SyncTime().CurrentTime(coms[0].Cns.Chr.ClockOffset()).String()))

	coms[0].ChRcvMsg[spos.MtBlockHeader] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, coms[0].Cns.RoundConsensus.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrBlock))

	// COMMITMENT_HASH
	cnsDta = spos.NewConsensusData(
		*coms[0].Cns.Data,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitmentHash,
		[]byte(coms[0].GetTime()))

	coms[0].ChRcvMsg[spos.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, coms[0].Cns.RoundConsensus.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrCommitmentHash))

	// BITMAP
	pks := make([][]byte, 0)

	for i := 0; i < len(coms[0].Cns.ConsensusGroup()); i++ {
		pks = append(pks, []byte(coms[0].Cns.ConsensusGroup()[i]))
	}

	cnsDta = spos.NewConsensusData(
		*coms[0].Cns.Data,
		pks,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtBitmap,
		[]byte(coms[0].GetTime()))

	coms[0].ChRcvMsg[spos.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < len(coms[0].Cns.ConsensusGroup()); i++ {
		assert.Equal(t, true, coms[0].Cns.RoundConsensus.GetJobDone(coms[0].Cns.ConsensusGroup()[i], spos.SrBitmap))
	}

	// COMMITMENT
	cnsDta = spos.NewConsensusData(
		*coms[0].Cns.Data,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitment,
		[]byte(coms[0].GetTime()))

	coms[0].ChRcvMsg[spos.MtCommitment] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, coms[0].Cns.RoundConsensus.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrCommitment))

	// SIGNATURE
	cnsDta = spos.NewConsensusData(
		*coms[0].Cns.Data,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtSignature,
		[]byte(coms[0].GetTime()))

	coms[0].ChRcvMsg[spos.MtSignature] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, coms[0].Cns.RoundConsensus.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrSignature))
}

func TestMessage_ReceivedBlock(t *testing.T) {
	coms := InitMessage()

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(coms[0].Cns.Chr.Round().TimeDuration()))

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtBlockBody,
		[]byte(coms[0].GetTime()))

	coms[0].Blk = &block.TxBlockBody{}

	r := coms[0].ReceivedBlockBody(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Blk = nil

	r = coms[0].ReceivedBlockBody(cnsDta)
	assert.Equal(t, true, r)

	cnsDta = spos.NewConsensusData(
		message,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtBlockHeader,
		[]byte(coms[0].GetTime()))

	coms[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)

	r = coms[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrBlock, spos.SsNotFinished)

	hdr.PrevHash = []byte("X")
	message, err = mock.MarshalizerMock{}.Marshal(hdr)
	assert.Nil(t, err)

	cnsDta.Data = message

	r = coms[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, false, r)

	hdr.PrevHash = []byte("")
	message, err = mock.MarshalizerMock{}.Marshal(hdr)
	assert.Nil(t, err)

	cnsDta.Data = message

	r = coms[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, coms[0].Cns.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrBlock))
}

func TestMessage_ReceivedCommitmentHash(t *testing.T) {
	coms := InitMessage()

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(coms[0].Cns.Chr.Round().TimeDuration()))

	dta := []byte("X")

	cnsDta := spos.NewConsensusData(
		dta,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitmentHash,
		[]byte(coms[0].GetTime()))

	for i := 0; i < coms[0].Cns.Threshold(spos.SrCommitmentHash); i++ {
		coms[0].Cns.RoundConsensus.SetJobDone(coms[0].Cns.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	r := coms[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, false, r)

	for i := 0; i < coms[0].Cns.Threshold(spos.SrCommitmentHash); i++ {
		coms[0].Cns.RoundConsensus.SetJobDone(coms[0].Cns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
	}

	coms[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)

	r = coms[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	r = coms[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, coms[0].Cns.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrCommitmentHash))
}

func TestMessage_ReceivedBitmap(t *testing.T) {
	coms := InitMessage()

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(coms[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*coms[0].Cns.Data,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitmentHash,
		[]byte(coms[0].GetTime()))

	coms[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)

	r := coms[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)

	pks := make([][]byte, 0)

	for i := 0; i < coms[0].Cns.Threshold(spos.SrBitmap)-1; i++ {
		pks = append(pks, []byte(coms[0].Cns.ConsensusGroup()[i]))
	}

	cnsDta.PubKeys = pks

	r = coms[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)
	assert.Equal(t, chronology.SubroundId(-1), coms[0].Cns.Chr.SelfSubround())

	cnsDta.PubKeys = append(cnsDta.PubKeys, []byte(coms[0].Cns.ConsensusGroup()[coms[0].Cns.Threshold(spos.SrBitmap)-1]))

	r = coms[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, coms[0].Cns.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrBitmap))

	for i := 0; i < coms[0].Cns.Threshold(spos.SrBitmap); i++ {
		assert.Equal(t, true, coms[0].Cns.RoundConsensus.GetJobDone(coms[0].Cns.ConsensusGroup()[i], spos.SrBitmap))
	}

	cnsDta.PubKeys = append(cnsDta.PubKeys, []byte("X"))

	r = coms[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)
	assert.Equal(t, chronology.SubroundId(-1), coms[0].Cns.Chr.SelfSubround())
}

func TestMessage_ReceivedCommitment(t *testing.T) {
	coms := InitMessage()

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(coms[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*coms[0].Cns.Data,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtCommitment,
		[]byte(coms[0].GetTime()))

	coms[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)

	r := coms[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)

	r = coms[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Cns.RoundConsensus.SetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrBitmap, true)

	r = coms[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, coms[0].Cns.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrCommitment))
}

func TestMessage_ReceivedSignature(t *testing.T) {
	coms := InitMessage()

	coms[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(coms[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*coms[0].Cns.Data,
		nil,
		[]byte(coms[0].Cns.ConsensusGroup()[1]),
		spos.MtSignature,
		[]byte(coms[0].GetTime()))

	coms[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r := coms[0].ReceivedSignature(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Cns.SetStatus(spos.SrSignature, spos.SsNotFinished)

	r = coms[0].ReceivedSignature(cnsDta)
	assert.Equal(t, false, r)

	coms[0].Cns.RoundConsensus.SetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrBitmap, true)

	r = coms[0].ReceivedSignature(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, coms[0].Cns.GetJobDone(coms[0].Cns.ConsensusGroup()[1], spos.SrSignature))
}

func TestMessage_CheckIfBlockIsValid(t *testing.T) {
	coms := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	hdr.PrevHash = []byte("X")

	r := coms[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.PrevHash = []byte("")

	r = coms[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 2

	r = coms[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 1
	coms[0].Blkc.CurrentBlockHeader = hdr

	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = []byte(coms[0].GetTime())

	r = coms[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")

	r = coms[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")

	r = coms[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 2

	r = coms[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)
}

func TestMessage_GetMessageTypeName(t *testing.T) {
	coms := InitMessage()

	r := coms[0].GetMessageTypeName(spos.MtBlockBody)
	assert.Equal(t, "<BLOCK_BODY>", r)

	r = coms[0].GetMessageTypeName(spos.MtBlockHeader)
	assert.Equal(t, "<BLOCK_HEADER>", r)

	r = coms[0].GetMessageTypeName(spos.MtCommitmentHash)
	assert.Equal(t, "<COMMITMENT_HASH>", r)

	r = coms[0].GetMessageTypeName(spos.MtBitmap)
	assert.Equal(t, "<BITMAP>", r)

	r = coms[0].GetMessageTypeName(spos.MtCommitment)
	assert.Equal(t, "<COMMITMENT>", r)

	r = coms[0].GetMessageTypeName(spos.MtSignature)
	assert.Equal(t, "<SIGNATURE>", r)

	r = coms[0].GetMessageTypeName(spos.MtUnknown)
	assert.Equal(t, "<UNKNOWN>", r)

	r = coms[0].GetMessageTypeName(spos.MessageType(-1))
	assert.Equal(t, "Undifined message type", r)
}

func TestConsensus_CheckConsensus(t *testing.T) {
	cns := InitConsensus()

	com := spos.NewCommunication(
		true,
		cns,
		nil,
		nil,
		mock.HasherMock{},
		mock.MarshalizerMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, cns, com)
	ok := cns.CheckStartRoundConsensus()
	assert.Equal(t, true, ok)

	ok = cns.CheckEndRoundConsensus()
	assert.Equal(t, true, ok)

	ok = cns.CheckSignatureConsensus()
	assert.Equal(t, true, ok)

	ok = cns.CheckSignatureConsensus()
	assert.Equal(t, true, ok)
}

func TestConsensus_CheckBlockConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrBlock, spos.SsNotFinished)

	ok := cns.CheckBlockConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrBlock))

	cns.SetJobDone("2", spos.SrBlock, true)

	ok = cns.CheckBlockConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrBlock))
}

func TestConsensus_CheckCommitmentHashConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	ok := cns.CheckCommitmentHashConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitmentHash))

	for i := 0; i < cns.Threshold(spos.SrCommitmentHash); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	ok = cns.CheckCommitmentHashConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitmentHash))

	cns.RoundConsensus.SetSelfId("2")

	cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	ok = cns.CheckCommitmentHashConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitmentHash))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	ok = cns.CheckCommitmentHashConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitmentHash))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrBitmap, false)
	}

	for i := 0; i < len(cns.RoundConsensus.ConsensusGroup()); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	ok = cns.CheckCommitmentHashConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitmentHash))
}

func TestConsensus_CheckBitmapConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)

	ok := cns.CheckBitmapConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrBitmap))

	for i := 1; i < len(cns.RoundConsensus.ConsensusGroup()); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	ok = cns.CheckBitmapConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrBitmap))

	cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[0], spos.SrCommitmentHash, true)

	ok = cns.CheckBitmapConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrBitmap))

	for i := 1; i < len(cns.RoundConsensus.ConsensusGroup()); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	cns.SetJobDone(cns.RoundConsensus.SelfId(), spos.SrBitmap, false)

	cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)

	ok = cns.CheckBitmapConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrBitmap))
}

func TestConsensus_CheckCommitmentConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)

	ok := cns.CheckCommitmentConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitment))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	for i := 1; i < len(cns.RoundConsensus.ConsensusGroup()); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrCommitment, true)
	}

	ok = cns.CheckCommitmentConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitment))

	cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[0], spos.SrCommitment, true)

	ok = cns.CheckCommitmentConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitment))
}

func TestConsensus_CheckSignatureConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrSignature, spos.SsNotFinished)

	ok := cns.CheckSignatureConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrSignature))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	for i := 1; i < cns.Threshold(spos.SrSignature); i++ {
		cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[i], spos.SrSignature, true)
	}

	ok = cns.CheckSignatureConsensus()
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrSignature))

	cns.SetJobDone(cns.RoundConsensus.ConsensusGroup()[0], spos.SrSignature, true)

	ok = cns.CheckSignatureConsensus()
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrSignature))
}
