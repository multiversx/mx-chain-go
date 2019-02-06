package bn_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
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

func InitSposWorker() *spos.Spos {
	eligibleList := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		9,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrSignature, false)
	}

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, 7)
	rthr.SetThreshold(bn.SrBitmap, 7)
	rthr.SetThreshold(bn.SrCommitment, 7)
	rthr.SetThreshold(bn.SrSignature, 7)

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrBlock, spos.SsFinished)
	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	rstatus.SetStatus(bn.SrBitmap, spos.SsFinished)
	rstatus.SetStatus(bn.SrCommitment, spos.SsFinished)
	rstatus.SetStatus(bn.SrSignature, spos.SsFinished)

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	chr.SetSelfSubround(0)

	sPoS := spos.NewSpos(
		nil,
		rcns,
		rthr,
		rstatus,
		chr,
	)

	return sPoS
}

func initSposWorkers() []*bn.Worker {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	eligibleList := CreateEligibleList(consensusGroupSize)
	// create instances
	var conWorkers []*bn.Worker

	for i := 0; i < consensusGroupSize; i++ {
		sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 1)
		cnWorker := initSposWorker(sPoS)
		GenerateSubRoundHandlers(roundDuration, sPoS, cnWorker)

		conWorkers = append(conWorkers, cnWorker)
	}

	return conWorkers
}

func initSposWorker(sPoS *spos.Spos) *bn.Worker {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	cnWorker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	cnWorker.Header = &block.Header{}
	cnWorker.SendMessage = SendMessage
	cnWorker.BroadcastBlockBody = BroadcastMessage
	cnWorker.BroadcastHeader = BroadcastMessage
	return cnWorker
}

func initSpos(
	genesisTime time.Time,
	roundDuration time.Duration,
	eligibleList []string,
	consensusGroupSize int,
	indexLeader int,
) *spos.Spos {

	chr := initChronology(genesisTime, roundDuration)
	rcns := initRoundConsensus(eligibleList, consensusGroupSize, indexLeader)
	rthr := initRoundThreshold(consensusGroupSize)
	rstatus := initRoundStatus()
	dta := []byte("X")

	chr.SetSelfSubround(0)

	sPoS := spos.NewSpos(
		dta,
		rcns,
		rthr,
		rstatus,
		chr,
	)

	return sPoS
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

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, PBFTThreshold)
	rthr.SetThreshold(bn.SrBitmap, PBFTThreshold)
	rthr.SetThreshold(bn.SrCommitment, PBFTThreshold)
	rthr.SetThreshold(bn.SrSignature, PBFTThreshold)

	return rthr
}

func initRoundStatus() *spos.RoundStatus {

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrBlock, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitment, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrSignature, spos.SsNotFinished)

	return rstatus
}

func initRoundConsensus(eligibleList []string, consensusGroupSize int, indexLeader int) *spos.RoundConsensus {
	rcns := spos.NewRoundConsensus(
		eligibleList,
		consensusGroupSize,
		eligibleList[indexLeader])

	rcns.SetConsensusGroup(eligibleList)

	for j := 0; j < len(rcns.ConsensusGroup()); j++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrSignature, false)
	}

	return rcns
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

	blockProcMock.ProcessAndCommitCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return nil
	}

	blockProcMock.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return nil
	}

	blockProcMock.CreateEmptyBlockBodyCalled = func(shardId uint32, round int32) *block.TxBlockBody {
		return &block.TxBlockBody{}
	}

	return blockProcMock
}

func initMultisigner() *mock.BelNevMock {
	multisigner := mock.NewMultiSigner()

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return []byte("commSecret"), []byte("commitment"), nil
	}

	multisigner.VerifySignatureShareMock = func(index uint16, sig []byte, bitmap []byte) error {
		return nil
	}

	multisigner.VerifyMock = func(bitmap []byte) error {
		return nil
	}

	multisigner.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return []byte("aggregatedSig"), nil
	}

	multisigner.AggregateCommitmentsMock = func(bitmap []byte) ([]byte, error) {
		return []byte("aggregatedCommitments"), nil
	}

	multisigner.CreateSignatureShareMock = func(bitmap []byte) ([]byte, error) {
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

	verifyMock := func(data []byte, signature []byte) error {
		return nil
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

func getEndTime(chr *chronology.Chronology, subroundID chronology.SubroundId) int64 {
	handlers := chr.SubroundHandlers()

	endTime := int64(0)

	for i := 0; i < len(handlers); i++ {
		if handlers[i].Current() == subroundID {
			endTime = handlers[i].EndTime()
			break
		}
	}

	return endTime
}

func GenerateSubRoundHandlers(roundDuration time.Duration, sPoS *spos.Spos, bnWorker *bn.Worker) {
	sPoS.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(bn.SrStartRound),
		chronology.SubroundId(bn.SrBlock),
		int64(roundDuration*5/100),
		bn.GetSubroundName(bn.SrStartRound),
		bnWorker.DoStartRoundJob,
		nil,
		func() bool { return true }))
	sPoS.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(roundDuration*25/100),
		bn.GetSubroundName(bn.SrBlock),
		bnWorker.DoBlockJob,
		bnWorker.ExtendBlock,
		bnWorker.CheckBlockConsensus))
	sPoS.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(bn.SrCommitmentHash),
		chronology.SubroundId(bn.SrBitmap),
		int64(roundDuration*40/100),
		bn.GetSubroundName(bn.SrCommitmentHash),
		bnWorker.DoCommitmentHashJob,
		bnWorker.ExtendCommitmentHash,
		bnWorker.CheckCommitmentHashConsensus))
	sPoS.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(bn.SrBitmap),
		chronology.SubroundId(bn.SrCommitment),
		int64(roundDuration*55/100),
		bn.GetSubroundName(bn.SrBitmap),
		bnWorker.DoBitmapJob,
		bnWorker.ExtendBitmap,
		bnWorker.CheckBitmapConsensus))
	sPoS.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(bn.SrCommitment),
		chronology.SubroundId(bn.SrSignature),
		int64(roundDuration*70/100),
		bn.GetSubroundName(bn.SrCommitment),
		bnWorker.DoCommitmentJob,
		bnWorker.ExtendCommitment,
		bnWorker.CheckCommitmentConsensus))
	sPoS.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(bn.SrSignature),
		chronology.SubroundId(bn.SrEndRound),
		int64(roundDuration*85/100),
		bn.GetSubroundName(bn.SrSignature),
		bnWorker.DoSignatureJob,
		bnWorker.ExtendSignature,
		bnWorker.CheckSignatureConsensus))
	sPoS.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(bn.SrEndRound),
		-1,
		int64(roundDuration*100/100),
		bn.GetSubroundName(bn.SrEndRound),
		bnWorker.DoEndRoundJob,
		bnWorker.ExtendEndRound,
		bnWorker.CheckEndRoundConsensus))
}

func CreateEligibleList(size int) []string {
	eligibleList := make([]string, 0)

	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, string(i+65))
	}

	return eligibleList
}

func TestWorker_NewWorkerConsensusNilShouldFail(t *testing.T) {
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroup, err := bn.NewWorker(
		nil,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		bootMock,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusGroup)
	assert.Equal(t, err, spos.ErrNilConsensus)
}

func TestWorker_NewWorkerBlockChainNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		nil,
		hasher,
		marshalizer,
		blkProc,
		bootMock,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilBlockChain)
}

func TestWorker_NewWorkerHasherNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		nil,
		marshalizer,
		blkProc,
		bootMock,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilHasher)
}

func TestWorker_NewWorkerMarshalizerNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		hasher,
		nil,
		blkProc,
		bootMock,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilMarshalizer)
}

func TestWorker_NewWorkerBlockProcessorNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		hasher,
		marshalizer,
		nil,
		bootMock,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilBlockProcessor)
}

func TestWorker_NewWorkerMultisigNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		bootMock,
		nil,
		keyGen,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestWorker_NewWorkerKeyGenNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		bootMock,
		multisig,
		nil,
		privKey,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilKeyGenerator)
}

func TestWorker_NewWorkerPrivKeyNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		bootMock,
		multisig,
		keyGen,
		nil,
		pubKey,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilPrivateKey)
}

func TestWorker_NewWorkerPubKeyNilFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		bootMock,
		multisig,
		keyGen,
		privKey,
		nil,
	)

	assert.Nil(t, consensusWorker)
	assert.Equal(t, err, spos.ErrNilPublicKey)
}

func TestWorker_NewWorkerShouldNotFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusWorker, err := bn.NewWorker(
		sPoS,
		blockChain,
		hasher,
		marshalizer,
		blkProc,
		bootMock,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	assert.NotNil(t, consensusWorker)
	assert.Nil(t, err)
}

func TestWorker_GetMessageTypeName(t *testing.T) {
	r := bn.GetMessageTypeName(bn.MtBlockBody)
	assert.Equal(t, "(BLOCK_BODY)", r)

	r = bn.GetMessageTypeName(bn.MtBlockHeader)
	assert.Equal(t, "(BLOCK_HEADER)", r)

	r = bn.GetMessageTypeName(bn.MtCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)

	r = bn.GetMessageTypeName(bn.MtBitmap)
	assert.Equal(t, "(BITMAP)", r)

	r = bn.GetMessageTypeName(bn.MtCommitment)
	assert.Equal(t, "(COMMITMENT)", r)

	r = bn.GetMessageTypeName(bn.MtSignature)
	assert.Equal(t, "(SIGNATURE)", r)

	r = bn.GetMessageTypeName(bn.MtUnknown)
	assert.Equal(t, "(UNKNOWN)", r)

	r = bn.GetMessageTypeName(bn.MessageType(-1))
	assert.Equal(t, "Undefined message type", r)
}

func TestWorker_GetSubroundName(t *testing.T) {
	r := bn.GetSubroundName(bn.SrStartRound)
	assert.Equal(t, "(START_ROUND)", r)

	r = bn.GetSubroundName(bn.SrBlock)
	assert.Equal(t, "(BLOCK)", r)

	r = bn.GetSubroundName(bn.SrCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)

	r = bn.GetSubroundName(bn.SrBitmap)
	assert.Equal(t, "(BITMAP)", r)

	r = bn.GetSubroundName(bn.SrCommitment)
	assert.Equal(t, "(COMMITMENT)", r)

	r = bn.GetSubroundName(bn.SrSignature)
	assert.Equal(t, "(SIGNATURE)", r)

	r = bn.GetSubroundName(bn.SrEndRound)
	assert.Equal(t, "(END_ROUND)", r)

	r = bn.GetSubroundName(bn.SrAdvance)
	assert.Equal(t, "(ADVANCE)", r)

	r = bn.GetSubroundName(chronology.SubroundId(-1))
	assert.Equal(t, "Undefined subround", r)
}
