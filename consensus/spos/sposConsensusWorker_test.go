package spos_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

type topicName string

const (
	consensusTopic topicName = "cns"
	// RoundTimeDuration defines the time duration in milliseconds of each round
	RoundTimeDuration = time.Duration(4000 * time.Millisecond)
)

func SendMessage(cnsDta *spos.ConsensusData) {
	fmt.Println(cnsDta.Signature)
}

func BroadcastMessage(msg []byte) {
	fmt.Println(msg)
}

func InitMessage() []*spos.SPOSConsensusWorker {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)
	// create instances
	var conWorkers []*spos.SPOSConsensusWorker

	for i := 0; i < consensusGroupSize; i++ {
		cns := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, i)
		cnWorker := initConsensusWorker(cns)
		GenerateSubRoundHandlers(roundDuration, cns, cnWorker)

		conWorkers = append(conWorkers, cnWorker)
	}

	return conWorkers
}

func initConsensusWorker(cns *spos.Consensus) *spos.SPOSConsensusWorker {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	cnWorker.SendMessage = SendMessage
	cnWorker.BroadcastBlockBody = BroadcastMessage
	cnWorker.BroadcastHeader = BroadcastMessage
	return cnWorker
}

func initConsensus(
	genesisTime time.Time,
	roundDuration time.Duration,
	consensusGroup []string,
	consensusGroupSize int,
	indexLeader int,
) *spos.Consensus {

	chr := initChronology(genesisTime, roundDuration)
	rndConsensus := initRoundConsensus(consensusGroup, indexLeader)
	rth := initRoundThreshold(consensusGroupSize)
	rnds := initRoundStatus()
	dta := []byte("X")

	cns := spos.NewConsensus(
		dta,
		rndConsensus,
		rth,
		rnds,
		chr,
	)
	return cns
}

func initChronology(genesisTime time.Time, roundDuration time.Duration) *chronology.Chronology {
	currentTime := genesisTime
	rnd := chronology.NewRound(genesisTime, currentTime, roundDuration)
	syncTime := ntp.NewSyncTime(roundDuration, nil)

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		syncTime,
	)

	return chr
}

func initRoundThreshold(consensusGroupSize int) *spos.RoundThreshold {
	PBFTThreshold := consensusGroupSize*2/3 + 1

	rth := spos.NewRoundThreshold()

	rth.SetThreshold(spos.SrBlock, 1)
	rth.SetThreshold(spos.SrCommitmentHash, PBFTThreshold)
	rth.SetThreshold(spos.SrBitmap, PBFTThreshold)
	rth.SetThreshold(spos.SrCommitment, PBFTThreshold)
	rth.SetThreshold(spos.SrSignature, PBFTThreshold)

	return rth
}

func initRoundStatus() *spos.RoundStatus {

	rnds := spos.NewRoundStatus()

	rnds.SetStatus(spos.SrBlock, spos.SsNotFinished)
	rnds.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)
	rnds.SetStatus(spos.SrBitmap, spos.SsNotFinished)
	rnds.SetStatus(spos.SrCommitment, spos.SsNotFinished)
	rnds.SetStatus(spos.SrSignature, spos.SsNotFinished)

	return rnds
}

func initRoundConsensus(consensusGroup []string, indexLeader int) *spos.RoundConsensus {
	rndConsensus := spos.NewRoundConsensus(
		consensusGroup,
		consensusGroup[indexLeader])

	for j := 0; j < len(rndConsensus.ConsensusGroup()); j++ {
		rndConsensus.SetJobDone(rndConsensus.ConsensusGroup()[j], spos.SrBlock, false)
		rndConsensus.SetJobDone(rndConsensus.ConsensusGroup()[j], spos.SrCommitmentHash, false)
		rndConsensus.SetJobDone(rndConsensus.ConsensusGroup()[j], spos.SrBitmap, false)
		rndConsensus.SetJobDone(rndConsensus.ConsensusGroup()[j], spos.SrCommitment, false)
		rndConsensus.SetJobDone(rndConsensus.ConsensusGroup()[j], spos.SrSignature, false)
	}
	return rndConsensus
}

func initMockBlockProcessor() *mock.BlockProcessorMock {
	blockProcMock := &mock.BlockProcessorMock{}

	blockProcMock.RemoveBlockTxsFromPoolCalled = func(*block.TxBlockBody) error { return nil }
	blockProcMock.CreateTxBlockCalled = func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
		return &block.TxBlockBody{}, nil
	}

	blockProcMock.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
		return nil
	}

	blockProcMock.RevertAccountStateCalled = func() {}

	blockProcMock.ProcessAndCommitCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
		return nil
	}

	blockProcMock.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
		return nil
	}

	return blockProcMock
}

func initMultisigner() *mock.BelNevMock {
	multisigner := mock.NewMultiSigner()

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return []byte("commSecret"), []byte("commitment"), nil
	}

	multisigner.VerifyPartialMock = func(index uint16, sig []byte) error {
		return nil
	}

	multisigner.VerifyMock = func() error {
		return nil
	}

	multisigner.AggregateSigsMock = func() ([]byte, error) {
		return []byte("aggregatedSig"), nil
	}

	multisigner.AggregateCommitmentsMock = func() ([]byte, error) {
		return []byte("aggregatedCommitments"), nil
	}

	multisigner.SignPartialMock = func() ([]byte, error) {
		return []byte("partialSign"), nil
	}

	return multisigner
}

func initSingleSigning() (*mock.KeyGenMock, *mock.PrivateKeyMock, *mock.PublicKeyMock) {
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

	return keyGenMock, privKeyMock, pubKeyMock
}

func GetTime(cw *spos.SPOSConsensusWorker) uint64 {
	return uint64(cw.Cns.Chr.SyncTime().CurrentTime(cw.Cns.Chr.ClockOffset()).Unix())
}

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

func TestNewConsensusWorker_ConsensusNilShouldFail(t *testing.T) {
	var consensus *spos.Consensus = nil
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusGroup, err := spos.NewConsensusWorker(
		consensus,
		nil,
		hasher,
		marshalizer,
		blkProc,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusGroup)
	assert.Equal(t, err, spos.ErrNilConsensus)
}

func TestNewConsensusWorker_BlockChainNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		nil,
		hasher,
		marshalizer,
		blkProc,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilBlockChain)
}

func TestNewConsensusWorker_HasherNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		blockChain,
		nil,
		marshalizer,
		blkProc,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilHasher)
}

func TestNewConsensusWorker_MarshalizerNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		blockChain,
		hasher,
		nil,
		blkProc,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilMarshalizer)
}

func TestNewConsensusWorker_BlockProcessorNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		blockChain,
		hasher,
		marshalizer,
		nil,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilBlockProcessor)
}

func TestNewConsensusWorker_MultisigNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		nil,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestNewConsensusWorker_KeyGenNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		multisig,
		nil,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilKeyGenerator)
}

func TestNewConsensusWorker_PrivKeyNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	pubKey := &mock.PublicKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		multisig,
		keyGen,
		nil,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilPrivateKey)
}

func TestNewConsensusWorker_PubKeyNilFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	consensus := initConsensus(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}

	consensusWorker, err := spos.NewConsensusWorker(
		consensus,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		multisig,
		keyGen,
		privKey,
		nil,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilPublicKey)
}

func TestNewConsensusData(t *testing.T) {
	cnsData := spos.NewConsensusData(
		nil,
		nil,
		nil,
		nil,
		spos.MtUnknown,
		0,
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
		nil,
		vld,
		nil,
		nil,
		nil,
	)

	msg2, _ := spos.NewConsensusWorker(
		cns,
		&blockchain.BlockChain{},
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

func TestSPOSConsensusWorker_DoEndRoundJobNotFinished(t *testing.T) {
	cnWorkers := InitMessage()

	cnWorkers[0].Hdr = &block.Header{}

	r := cnWorkers[0].DoEndRoundJob()
	assert.False(t, r)
}

func TestSPOSConsensusWorker_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	multisigner.AggregateSigsMock = func() ([]byte, error) {
		return nil, crypto.ErrNilHasher
	}

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	cnWorker.SendMessage = SendMessage
	cnWorker.BroadcastBlockBody = BroadcastMessage
	cnWorker.BroadcastHeader = BroadcastMessage

	cnWorker.Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, cns, cnWorker)
	cnWorker.Hdr = &block.Header{}

	r := cnWorker.DoEndRoundJob()
	assert.False(t, r)
}

func TestSPOSConsensusWorker_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker := initConsensusWorker(cns)

	blProcMock := initMockBlockProcessor()

	blProcMock.CommitBlockCalled = func(
		blockChain *blockchain.BlockChain,
		header *block.Header,
		block *block.TxBlockBody,
	) error {
		return blockchain.ErrHeaderUnitNil
	}

	cnWorker.BlockProcessor = blProcMock

	cnWorker.Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, cns, cnWorker)
	cnWorker.Hdr = &block.Header{}

	r := cnWorker.DoEndRoundJob()
	assert.False(t, r)
}

func TestSPOSConsensusWorker_DoEndRoundJobErrRemBlockTxOK(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker := initConsensusWorker(cns)

	blProcMock := initMockBlockProcessor()

	blProcMock.RemoveBlockTxsFromPoolCalled = func(body *block.TxBlockBody) error {
		return process.ErrNilBlockBodyPool
	}

	cnWorker.BlockProcessor = blProcMock

	cnWorker.Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, cns, cnWorker)
	cnWorker.Hdr = &block.Header{}

	r := cnWorker.DoEndRoundJob()
	assert.True(t, r)
}

func TestSPOSConsensusWorker_DoEndRoundJobErrBroadcastTxBlockBodyOK(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker := initConsensusWorker(cns)

	cnWorker.Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	cnWorker.BroadcastBlockBody = nil

	GenerateSubRoundHandlers(roundDuration, cns, cnWorker)
	cnWorker.Hdr = &block.Header{}

	r := cnWorker.DoEndRoundJob()
	assert.True(t, r)
}

func TestSPOSConsensusWorker_DoEndRoundJobErrBroadcastHeaderOK(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker := initConsensusWorker(cns)

	cnWorker.Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	cnWorker.BroadcastHeader = nil

	GenerateSubRoundHandlers(roundDuration, cns, cnWorker)
	cnWorker.Hdr = &block.Header{}

	r := cnWorker.DoEndRoundJob()
	assert.True(t, r)
}

func TestSPOSConsensusWorker_DoEndRoundJobAllOK(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker := initConsensusWorker(cns)

	cnWorker.Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrBitmap, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitment, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, cns, cnWorker)
	cnWorker.Hdr = &block.Header{}

	r := cnWorker.DoEndRoundJob()
	assert.True(t, r)
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
	cnWorkers[0].Cns.Data = dta

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.Equal(t, true, r)
}

func TestSPOSConsensusWorker_DoCommitmentHashJobErrCreateCommitmentShouldFail(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, crypto.ErrNilHasher
	}

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	cnWorker.SendMessage = SendMessage
	cnWorker.BroadcastBlockBody = BroadcastMessage
	cnWorker.BroadcastHeader = BroadcastMessage

	cnWorker.Cns.SetStatus(spos.SrBlock, spos.SsFinished)
	cnWorker.Cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	done := cnWorker.DoCommitmentHashJob()

	assert.False(t, done)
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
	cnWorkers[0].Cns.Data = dta
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
	cnWorkers[0].Cns.Data = dta

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
	cnWorkers[0].Cns.Data = dta

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
		cnWorkers[0].GetTime(),
		0,
	)

	cnWorkers[0].SendMessage = nil
	r := cnWorkers[0].SendConsensusMessage(cnsDta)
	assert.Equal(t, false, r)

	cnWorkers[0].SendMessage = SendMessage
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
		cnWorkers[0].GetTime(),
		0,
	)

	cnWorkers[0].ReceivedMessage(string(consensusTopic), cnsDta, nil)
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
		cnWorkers[0].GetTime(),
		0,
	)

	cnWorkers[0].ReceivedMessage(string(consensusTopic), cnsDta, nil)
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

	// BLOCK BODY
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		GetTime(cnWorkers[0]),
		0,
	)

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

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = GetTime(cnWorkers[0])

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		GetTime(cnWorkers[0]),
		0,
	)

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
		cnWorkers[0].Cns.Data,
		commitmentHash,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime(),
		0,
	)

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
		cnWorkers[0].Cns.Data,
		bitmap,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBitmap,
		cnWorkers[0].GetTime(),
		0,
	)

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
		cnWorkers[0].Cns.Data,
		commHash,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime(),
		0,
	)

	cnWorkers[0].ChRcvMsg[spos.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)

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
		cnWorkers[0].Cns.Data,
		bitmap,
		[]byte(cnsGroup[1]),
		[]byte("sig"),
		spos.MtBitmap,
		cnWorkers[0].GetTime(),
		0,
	)

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
		cnWorkers[0].GetTime(),
		0,
	)

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
		cnWorkers[0].GetTime(),
		0,
	)

	r = cnWorkers[0].ReceivedBlockHeader(cnsDta)
	assert.Equal(t, true, r)
}

func TestSPOSConsensusWorker_ReceivedBlockBodyHeaderReceivedJobDone(t *testing.T) {
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
		cnWorkers[0].GetTime(),
		0,
	)

	cnWorkers[0].Hdr = &block.Header{}

	r := cnWorkers[0].ReceivedBlockBody(cnsDta)
	assert.True(t, r)
}

func TestSPOSConsensusWorker_ReceivedBlockBodyHeaderReceivedErrProcessBlockShouldFail(t *testing.T) {
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
		cnWorkers[0].GetTime(),
		0,
	)

	cnWorkers[0].Hdr = &block.Header{}

	blProcMock := initMockBlockProcessor()

	blProcMock.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
		return process.ErrNilPreviousBlockHash
	}

	cnWorkers[0].BlockProcessor = blProcMock

	r := cnWorkers[0].ReceivedBlockBody(cnsDta)
	assert.False(t, r)
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
		cnWorkers[0].GetTime(),
		0,
	)

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
		cnWorkers[0].Cns.Data,
		[]byte("commHash"),
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime(),
		0,
	)

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
		cnWorkers[0].Cns.Data,
		commHash,
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorkers[0].GetTime(),
		0)

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
		cnWorkers[0].Cns.Data,
		nil,
		[]byte(cnWorkers[0].Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtSignature,
		cnWorkers[0].GetTime(),
		0,
	)

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
