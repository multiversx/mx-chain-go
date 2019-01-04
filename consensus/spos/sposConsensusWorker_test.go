package spos_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
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
	var conWorkers []*spos.SPOSConsensusWorker

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

		toByteArrayMock := func() ([]byte, error) {
			return []byte("byteArray"), nil
		}

		signMock := func(message []byte) ([]byte, error) {
			return []byte("signature"), nil
		}

		verifyMock := func(data []byte, signature []byte) (bool, error) {
			return true, nil
		}

		privKeyMock := &mock.PrivateKeyMock{
			SignMock:        signMock,
			ToByteArrayMock: toByteArrayMock,
		}

		pubKeyMock := &mock.PublicKeyMock{
			ToByteArrayMock: toByteArrayMock,
			VerifyMock:      verifyMock,
		}

		privKeyFromByteArr := func(b []byte) (crypto.PrivateKey, error) {
			return privKeyMock, nil
		}

		pubKeyFromByteArr := func(b []byte) (crypto.PublicKey, error) {
			return pubKeyMock, nil
		}

		keyGenMock := &mock.KeyGenMock{
			PrivateKeyFromByteArrayMock: privKeyFromByteArr,
			PublicKeyFromByteArrayMock:  pubKeyFromByteArr,
		}

		multisigner := mock.NewMultiSigner()

		multisigner.VerifyPartialMock = func(index uint16, sig []byte) error {
			return nil
		}

		multisigner.VerifyMock = func() error {
			return nil
		}

		cnWorker, _ := spos.NewConsensusWorker(
			cns,
			&blkc,
			mock.HasherMock{},
			mock.MarshalizerMock{},
			&mock.BlockProcessorMock{},
			multisigner,
			keyGenMock,
			privKeyMock,
			pubKeyMock)

		cnWorker.OnSendMessage = SendMessage
		cnWorker.OnBroadcastBlockBody = SendMessage
		cnWorker.OnBroadcastHeader = SendMessage
		GenerateSubRoundHandlers(roundDuration, cns, cnWorker)

		conWorkers = append(conWorkers, cnWorker)
	}

	conWorkers[0].BlockProcessor = &mock.BlockProcessorMock{
		RemoveBlockTxsFromPoolCalled: func(*block.TxBlockBody) error { return nil },
		CreateTxBlockCalled: func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
			return &block.TxBlockBody{}, nil
		},
		CommitBlockCalled: func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
			return nil
		},
		RevertAccountStateCalled: func() {},
		ProcessAndCommitCalled: func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
			return nil
		},
		ProcessBlockCalled: func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
			return nil
		},
	}

	return conWorkers
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

func GenerateSubRoundHandlers(roundDuration time.Duration, cns *spos.Consensus, cnWorker *spos.SPOSConsensusWorker) {
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrStartRound),
		chronology.SubroundId(spos.SrBlock),
		int64(roundDuration*5/100),
		cns.GetSubroundName(spos.SrStartRound),
		cnWorker.DoStartRoundJob,
		nil,
		func() bool { return true }))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBlock),
		chronology.SubroundId(spos.SrCommitmentHash),
		int64(roundDuration*25/100),
		cns.GetSubroundName(spos.SrBlock),
		cnWorker.DoBlockJob, cnWorker.ExtendBlock,
		cns.CheckBlockConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitmentHash),
		chronology.SubroundId(spos.SrBitmap),
		int64(roundDuration*40/100),
		cns.GetSubroundName(spos.SrCommitmentHash),
		cnWorker.DoCommitmentHashJob,
		cnWorker.ExtendCommitmentHash,
		cns.CheckCommitmentHashConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBitmap),
		chronology.SubroundId(spos.SrCommitment),
		int64(roundDuration*55/100),
		cns.GetSubroundName(spos.SrBitmap),
		cnWorker.DoBitmapJob,
		cnWorker.ExtendBitmap,
		cns.CheckBitmapConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitment),
		chronology.SubroundId(spos.SrSignature),
		int64(roundDuration*70/100),
		cns.GetSubroundName(spos.SrCommitment),
		cnWorker.DoCommitmentJob,
		cnWorker.ExtendCommitment,
		cns.CheckCommitmentConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrSignature),
		chronology.SubroundId(spos.SrEndRound),
		int64(roundDuration*85/100),
		cns.GetSubroundName(spos.SrSignature),
		cnWorker.DoSignatureJob,
		cnWorker.ExtendSignature,
		cns.CheckSignatureConsensus))
	cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrEndRound),
		-1,
		int64(roundDuration*100/100),
		cns.GetSubroundName(spos.SrEndRound),
		cnWorker.DoEndRoundJob,
		cnWorker.ExtendEndRound,
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
		nil,
		spos.MtUnknown,
		0)

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

	cnWorker, _ := spos.NewConsensusWorker(
		nil,
		nil,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	assert.Equal(t, 0, cap(cnWorker.ChRcvMsg[spos.MtBlockHeader]))

	msg2, _ := spos.NewConsensusWorker(
		cns,
		nil,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	assert.Equal(t, len(cns.RoundConsensus.ConsensusGroup()), cap(msg2.ChRcvMsg[spos.MtBlockHeader]))
}

func TestMessage_StartRound(t *testing.T) {
	cnWorkers := InitMessage()

	r := cnWorkers[0].DoStartRoundJob()
	assert.Equal(t, true, r)
}

func TestMessage_EndRound(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Hdr = &block.Header{}

	r := cnWorkers[0].DoEndRoundJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r = cnWorkers[0].DoEndRoundJob()
	assert.Equal(t, true, r)

	cnWorkers[0].Cns.RoundConsensus.SetSelfPubKey(cnWorkers[0].Cns.RoundConsensus.ConsensusGroup()[1])

	r = cnWorkers[0].DoEndRoundJob()
	assert.Equal(t, 2, cnWorkers[0].RoundsWithBlock)
	assert.Equal(t, true, r)
}

func TestMessage_SendBlock(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].Cns.Chr.Round().TimeDuration()))

	r := cnWorkers[0].DoBlockJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now())
	cnWorkers[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)

	r = cnWorkers[0].DoBlockJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrBlock, spos.SsNotFinished)
	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBlock, true)

	r = cnWorkers[0].DoBlockJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBlock, false)
	cnWorkers[0].Cns.RoundConsensus.SetSelfPubKey(cnWorkers[0].Cns.RoundConsensus.ConsensusGroup()[1])

	r = cnWorkers[0].DoBlockJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.RoundConsensus.SetSelfPubKey(cnWorkers[0].Cns.RoundConsensus.ConsensusGroup()[0])

	r = cnWorkers[0].DoBlockJob()
	assert.Equal(t, true, r)
	assert.Equal(t, uint64(1), cnWorkers[0].Hdr.Nonce)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBlock, false)
	cnWorkers[0].Blkc.CurrentBlockHeader = cnWorkers[0].Hdr

	r = cnWorkers[0].DoBlockJob()
	assert.Equal(t, true, r)
	assert.Equal(t, uint64(2), cnWorkers[0].Hdr.Nonce)
}

func TestMessage_SendCommitmentHash(t *testing.T) {
	cnWorkers := InitMessage()

	r := cnWorkers[0].DoCommitmentHashJob()
	assert.Equal(t, true, r)

	cnWorkers[0].Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)
	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrCommitmentHash, true)

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrCommitmentHash, false)
	cnWorkers[0].Cns.Data = nil

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	cnWorkers[0].Cns.Data = &dta

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.Equal(t, true, r)
}

func TestMessage_SendBitmap(t *testing.T) {
	cnWorkers := InitMessage()

	r := cnWorkers[0].DoBitmapJob()
	assert.Equal(t, true, r)

	cnWorkers[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)

	r = cnWorkers[0].DoBitmapJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)
	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBitmap, true)

	r = cnWorkers[0].DoBitmapJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBitmap, false)
	cnWorkers[0].Cns.RoundConsensus.SetSelfPubKey(cnWorkers[0].Cns.RoundConsensus.ConsensusGroup()[1])

	r = cnWorkers[0].DoBitmapJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.RoundConsensus.SetSelfPubKey(cnWorkers[0].Cns.RoundConsensus.ConsensusGroup()[0])
	cnWorkers[0].Cns.Data = nil

	r = cnWorkers[0].DoBitmapJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	cnWorkers[0].Cns.Data = &dta
	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrCommitmentHash, true)

	r = cnWorkers[0].DoBitmapJob()
	assert.Equal(t, true, r)
	assert.Equal(t, true, cnWorkers[0].Cns.GetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBitmap))
}

func TestMessage_SendCommitment(t *testing.T) {
	cnWorkers := InitMessage()

	r := cnWorkers[0].DoCommitmentJob()
	assert.Equal(t, true, r)

	cnWorkers[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)

	r = cnWorkers[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)
	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrCommitment, true)

	r = cnWorkers[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrCommitment, false)

	r = cnWorkers[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBitmap, true)
	cnWorkers[0].Cns.Data = nil

	r = cnWorkers[0].DoCommitmentJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	cnWorkers[0].Cns.Data = &dta

	r = cnWorkers[0].DoCommitmentJob()
	assert.Equal(t, true, r)
}

func TestMessage_SendSignature(t *testing.T) {
	cnWorkers := InitMessage()

	r := cnWorkers[0].DoSignatureJob()
	assert.Equal(t, true, r)

	cnWorkers[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorkers[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r = cnWorkers[0].DoSignatureJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrSignature, spos.SsNotFinished)
	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrSignature, true)

	r = cnWorkers[0].DoSignatureJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrSignature, false)

	r = cnWorkers[0].DoSignatureJob()
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.SelfPubKey(), spos.SrBitmap, true)
	cnWorkers[0].Cns.Data = nil

	r = cnWorkers[0].DoSignatureJob()
	assert.Equal(t, false, r)

	dta := []byte("X")
	cnWorkers[0].Cns.Data = &dta

	r = cnWorkers[0].DoSignatureJob()
	assert.Equal(t, true, r)
}

func TestMessage_BroadcastMessage(t *testing.T) {
	cnWorkers := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].GetTime()

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorkers[0].Cns.SelfPubKey()),
		[]byte("sig"),
		spos.MtBlockHeader,
		cnWorkers[0].GetTime())

	cnWorkers[0].OnSendMessage = nil
	r := cnWorkers[0].SendConsensusMessage(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].OnSendMessage = SendMessage
	r = cnWorkers[0].SendConsensusMessage(cnsDta)
	assert.Equal(t, true, r)
}

func TestMessage_ExtendBlock(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].ExtendBlock()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].Cns.Status(spos.SrBlock))
}

func TestMessage_ExtendCommitmentHash(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].Cns.Status(spos.SrCommitmentHash))

	for i := 0; i < len(cnWorkers[0].Cns.RoundConsensus.ConsensusGroup()); i++ {
		cnWorkers[0].Cns.SetJobDone(cnWorkers[0].Cns.RoundConsensus.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	cnWorkers[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].Cns.Status(spos.SrCommitmentHash))
}

func TestMessage_ExtendBitmap(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].ExtendBitmap()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].Cns.Status(spos.SrBitmap))
}

func TestMessage_ExtendCommitment(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].ExtendCommitment()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].Cns.Status(spos.SrCommitment))
}

func TestMessage_ExtendSignature(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].ExtendSignature()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].Cns.Status(spos.SrSignature))
}

func TestMessage_ExtendEndRound(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].ExtendEndRound()
}

func TestMessage_ReceivedMessageTxBlockBody(t *testing.T) {
	cnWorkers := InitMessage()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorkers[0].Cns.SelfPubKey()),
		[]byte("sig"),
		spos.MtBlockBody,
		uint64(cnWorkers[0].Cns.Chr.SyncTime().CurrentTime(cnWorkers[0].Cns.Chr.ClockOffset()).Unix()))

	cnWorkers[0].ReceivedMessage(cnsDta)
}

func TestMessage_ReceivedMessageUnknown(t *testing.T) {
	cnWorkers := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].GetTime()

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))
	message, _ = mock.MarshalizerMock{}.Marshal(hdr)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorkers[0].Cns.SelfPubKey()),
		[]byte("sig"),
		spos.MtUnknown,
		cnWorkers[0].GetTime())

	cnWorkers[0].ReceivedMessage(cnsDta)
}

func TestMessage_DecodeBlockBody(t *testing.T) {
	cnWorkers := InitMessage()

	blk := &block.TxBlockBody{}

	mblks := make([]block.MiniBlock, 0)
	mblks = append(mblks, block.MiniBlock{ShardID: 69})
	blk.MiniBlocks = mblks

	message, err := mock.MarshalizerMock{}.Marshal(blk)

	assert.Nil(t, err)

	dcdBlk := cnWorkers[0].DecodeBlockBody(nil)

	assert.Nil(t, dcdBlk)

	dcdBlk = cnWorkers[0].DecodeBlockBody(&message)

	assert.Equal(t, blk, dcdBlk)
	assert.Equal(t, uint32(69), dcdBlk.MiniBlocks[0].ShardID)
}

func TestMessage_DecodeBlockHeader(t *testing.T) {
	cnWorkers := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].GetTime()
	hdr.Signature = []byte(cnWorkers[0].Cns.SelfPubKey())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	dcdHdr := cnWorkers[0].DecodeBlockHeader(nil)

	assert.Nil(t, dcdHdr)

	dcdHdr = cnWorkers[0].DecodeBlockHeader(&message)

	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte(cnWorkers[0].Cns.SelfPubKey()), dcdHdr.Signature)
}

func TestMessage_CheckChannelTxBlockBody(t *testing.T) {
	cnWorkers := InitMessage()
	round := cnWorkers[0].Cns.Chr.Round()
	roundDuration := round.TimeDuration()
	round.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	cnsGroup := cnWorkers[0].Cns.ConsensusGroup()
	rndCns := cnWorkers[0].Cns.RoundConsensus
	chr := cnWorkers[0].Cns.Chr

	// BLOCK BODY
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		uint64(chr.SyncTime().CurrentTime(chr.ClockOffset()).Unix()))

	cnWorkers[0].ChRcvMsg[spos.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isBlockJobDone := rndCns.GetJobDone(cnsGroup[1], spos.SrBlock)

	// Not done since header is missing
	assert.False(t, isBlockJobDone)
}

func TestMessage_CheckChannelBlockHeader(t *testing.T) {
	cnWorkers := InitMessage()

	round := cnWorkers[0].Cns.Chr.Round()
	roundDuration := round.TimeDuration()
	round.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	cnsGroup := cnWorkers[0].Cns.ConsensusGroup()
	rndCns := cnWorkers[0].Cns.RoundConsensus
	chr := cnWorkers[0].Cns.Chr

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(chr.SyncTime().CurrentTime(chr.ClockOffset()).Unix())

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		uint64(chr.SyncTime().CurrentTime(chr.ClockOffset()).Unix()))

	cnWorkers[0].ChRcvMsg[spos.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	cnsDta.MsgType = spos.MtBlockHeader
	cnWorkers[0].ChRcvMsg[spos.MtBlockHeader] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	isBlockJobDone := rndCns.GetJobDone(cnsGroup[1], spos.SrBlock)

	assert.True(t, isBlockJobDone)
}

func TestConsensus_CheckChannelsCommitmentHash(t *testing.T) {
	cnWorkers := InitMessage()

	round := cnWorkers[0].Cns.Chr.Round()
	roundDuration := round.TimeDuration()
	round.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rndCns := cnWorkers[0].Cns.RoundConsensus
	cnsGroup := cnWorkers[0].Cns.ConsensusGroup()

	commitmentHash := []byte("commitmentHash")

	// COMMITMENT_HASH
	cnsDta := spos.NewConsensusData(
		*cnWorkers[0].Cns.Data,
		commitmentHash,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime())

	cnWorkers[0].ChRcvMsg[spos.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isCommitmentHashJobDone := rndCns.GetJobDone(cnsGroup[1], spos.SrCommitmentHash)

	assert.True(t, isCommitmentHashJobDone)
}

func TestConsensus_CheckChannelsBitmap(t *testing.T) {
	cnWorkers := InitMessage()

	round := cnWorkers[0].Cns.Chr.Round()
	roundDuration := round.TimeDuration()
	round.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rndCns := cnWorkers[0].Cns.RoundConsensus
	cnsGroup := cnWorkers[0].Cns.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	// BITMAP
	cnsDta := spos.NewConsensusData(
		*cnWorkers[0].Cns.Data,
		bitmap,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBitmap,
		cnWorkers[0].GetTime())

	cnWorkers[0].ChRcvMsg[spos.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < len(cnsGroup); i++ {
		isBitmapJobDone := rndCns.GetJobDone(cnsGroup[i], spos.SrBitmap)
		assert.True(t, isBitmapJobDone)
	}
}

func TestMessage_CheckChannelsCommitment(t *testing.T) {
	cnWorkers := InitMessage()

	round := cnWorkers[0].Cns.Chr.Round()
	roundDuration := round.TimeDuration()
	round.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rndCns := cnWorkers[0].Cns.RoundConsensus
	cnsGroup := cnWorkers[0].Cns.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	commHash := mock.HasherMock{}.Compute("commitment")

	// Commitment Hash
	cnsDta := spos.NewConsensusData(
		*cnWorkers[0].Cns.Data,
		commHash,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime())

	cnWorkers[0].ChRcvMsg[spos.MtCommitmentHash] <- cnsDta
	time.Sleep(10*time.Millisecond)

	// Bitmap
	cnsDta.MsgType = spos.MtBitmap
	cnsDta.SubRoundData = bitmap
	cnWorkers[0].ChRcvMsg[spos.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	// Commitment
	cnsDta.MsgType = spos.MtCommitment
	cnsDta.SubRoundData = []byte("commitment")
	cnWorkers[0].ChRcvMsg[spos.MtCommitment] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	isCommitmentJobDone := rndCns.GetJobDone(cnsGroup[1], spos.SrCommitment)

	assert.True(t, isCommitmentJobDone)
}

func TestMessage_CheckChannelsSignature(t *testing.T) {
	cnWorkers := InitMessage()

	round := cnWorkers[0].Cns.Chr.Round()
	roundDuration := round.TimeDuration()
	round.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rndCns := cnWorkers[0].Cns.RoundConsensus
	cnsGroup := cnWorkers[0].Cns.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	// Bitmap
	cnsDta := spos.NewConsensusData(
		*cnWorkers[0].Cns.Data,
		bitmap,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBitmap,
		cnWorkers[0].GetTime())

	cnWorkers[0].ChRcvMsg[spos.MtBitmap] <- cnsDta

	// Signature
	cnsDta.MsgType = spos.MtSignature
	cnsDta.SubRoundData = []byte("signature")
	time.Sleep(10 * time.Millisecond)

	cnWorkers[0].ChRcvMsg[spos.MtSignature] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isSigJobDone := rndCns.GetJobDone(cnsGroup[1], spos.SrSignature)

	assert.True(t, isSigJobDone)
}

func TestMessage_ReceivedBlock(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].Cns.Chr.Round().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		cnWorkers[0].GetTime())

	cnWorkers[0].Blk = &block.TxBlockBody{}

	r := cnWorkers[0].ReceivedBlockBody(cnsDta)
	assert.Equal(t, false, r)

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(blBodyStr))

	hdrStr, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash := mock.HasherMock{}.Compute(string(hdrStr))

	cnsDta = spos.NewConsensusData(
		hdrHash,
		hdrStr,
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockHeader,
		cnWorkers[0].GetTime())

	r = cnWorkers[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, true, r)
}

func TestMessage_ReceivedCommitmentHash(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].Cns.Chr.Round().TimeDuration()))

	dta := []byte("X")

	cnsDta := spos.NewConsensusData(
		dta,
		nil,
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		nil,
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime())

	for i := 0; i < cnWorkers[0].Cns.Threshold(spos.SrCommitmentHash); i++ {
		cnWorkers[0].Cns.RoundConsensus.SetJobDone(cnWorkers[0].Cns.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	r := cnWorkers[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, false, r)

	for i := 0; i < cnWorkers[0].Cns.Threshold(spos.SrCommitmentHash); i++ {
		cnWorkers[0].Cns.RoundConsensus.SetJobDone(cnWorkers[0].Cns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
	}

	cnWorkers[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)

	r = cnWorkers[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	r = cnWorkers[0].ReceivedCommitmentHash(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, cnWorkers[0].Cns.GetJobDone(cnWorkers[0].Cns.ConsensusGroup()[1], spos.SrCommitmentHash))
}

func TestMessage_ReceivedBitmap(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*cnWorkers[0].Cns.Data,
		[]byte("commHash"),
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime())

	cnWorkers[0].Cns.SetStatus(spos.SrBitmap, spos.SsFinished)

	r := cnWorkers[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)

	bitmap := make([]byte, 3)

	cnGroup := cnWorkers[0].Cns.ConsensusGroup()

	// fill ony few of the signers in bitmap
	for i := 0; i < 5; i++ {
		bitmap[i/8] |= 1 << uint16(i%8)
	}

	cnsDta.SubRoundData = bitmap

	r = cnWorkers[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, false, r)

	//fill the rest
	for i := 5; i < len(cnGroup); i++ {
		bitmap[i/8] |= 1 << uint16(i%8)
	}

	cnsDta.SubRoundData = bitmap

	r = cnWorkers[0].ReceivedBitmap(cnsDta)
	assert.Equal(t, true, r)
}

func TestMessage_ReceivedCommitment(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].Cns.Chr.Round().TimeDuration()))

	commitment := []byte("commitment")
	commHash := mock.HasherMock{}.Compute(string(commitment))

	cnsDta := spos.NewConsensusData(
		*cnWorkers[0].Cns.Data,
		commHash,
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime())

	r := cnWorkers[0].ReceivedCommitmentHash(cnsDta)

	cnsDta.MsgType = spos.MtCommitment
	cnsDta.SubRoundData = commitment

	cnWorkers[0].Cns.SetStatus(spos.SrCommitment, spos.SsFinished)

	r = cnWorkers[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)

	r = cnWorkers[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.RoundConsensus.SetJobDone(cnWorkers[0].Cns.ConsensusGroup()[1], spos.SrBitmap, true)

	r = cnWorkers[0].ReceivedCommitment(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, cnWorkers[0].Cns.GetJobDone(cnWorkers[0].Cns.ConsensusGroup()[1], spos.SrCommitment))
}

func TestMessage_ReceivedSignature(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Cns.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].Cns.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		*cnWorkers[0].Cns.Data,
		nil,
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtSignature,
		cnWorkers[0].GetTime())

	cnWorkers[0].Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	r := cnWorkers[0].ReceivedSignature(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.SetStatus(spos.SrSignature, spos.SsNotFinished)

	r = cnWorkers[0].ReceivedSignature(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].Cns.RoundConsensus.SetJobDone(cnWorkers[0].Cns.ConsensusGroup()[1], spos.SrBitmap, true)

	r = cnWorkers[0].ReceivedSignature(cnsDta)
	assert.Equal(t, true, r)
	assert.Equal(t, true, cnWorkers[0].Cns.GetJobDone(cnWorkers[0].Cns.ConsensusGroup()[1], spos.SrSignature))
}

func TestMessage_CheckIfBlockIsValid(t *testing.T) {
	cnWorkers := InitMessage()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].GetTime()

	hdr.PrevHash = []byte("X")

	r := cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.PrevHash = []byte("")

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 2

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 1
	cnWorkers[0].Blkc.CurrentBlockHeader = hdr

	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].GetTime()

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, false, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)

	hdr.Nonce = 2

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.Equal(t, true, r)
}

func TestMessage_GetMessageTypeName(t *testing.T) {
	cnWorkers := InitMessage()

	r := cnWorkers[0].GetMessageTypeName(spos.MtBlockBody)
	assert.Equal(t, "<BLOCK_BODY>", r)

	r = cnWorkers[0].GetMessageTypeName(spos.MtBlockHeader)
	assert.Equal(t, "<BLOCK_HEADER>", r)

	r = cnWorkers[0].GetMessageTypeName(spos.MtCommitmentHash)
	assert.Equal(t, "<COMMITMENT_HASH>", r)

	r = cnWorkers[0].GetMessageTypeName(spos.MtBitmap)
	assert.Equal(t, "<BITMAP>", r)

	r = cnWorkers[0].GetMessageTypeName(spos.MtCommitment)
	assert.Equal(t, "<COMMITMENT>", r)

	r = cnWorkers[0].GetMessageTypeName(spos.MtSignature)
	assert.Equal(t, "<SIGNATURE>", r)

	r = cnWorkers[0].GetMessageTypeName(spos.MtUnknown)
	assert.Equal(t, "<UNKNOWN>", r)

	r = cnWorkers[0].GetMessageTypeName(spos.MessageType(-1))
	assert.Equal(t, "Undifined message type", r)
}

func TestConsensus_CheckConsensus(t *testing.T) {
	cns := InitConsensus()

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
		nil,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, cns, cnWorker)
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

	cns.RoundConsensus.SetSelfPubKey("2")

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

	cns.SetJobDone(cns.RoundConsensus.SelfPubKey(), spos.SrBitmap, false)

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
