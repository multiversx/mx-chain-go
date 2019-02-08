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
)

func sendMessage(cnsDta *spos.ConsensusData) {
	fmt.Println(cnsDta.Signature)
}

func broadcastMessage(msg []byte) {
	fmt.Println(msg)
}

func initWorker() *bn.Worker {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := createEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		1,
	)

	wrk, _ := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	wrk.SendMessage = sendMessage
	wrk.BroadcastBlockBody = broadcastMessage
	wrk.BroadcastHeader = broadcastMessage

	return wrk
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

func createEligibleList(size int) []string {
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilConsensus)
}

func TestWorker_NewWorkerBlockChainNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	eligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilBlockChain)
}

func TestWorker_NewWorkerHasherNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	eligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilHasher)
}

func TestWorker_NewWorkerMarshalizerNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	eligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilMarshalizer)
}

func TestWorker_NewWorkerBlockProcessorNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	eligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilBlockProcessor)
}

func TestWorker_NewWorkerMultisigNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	eligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestWorker_NewWorkerKeyGenNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	eligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilKeyGenerator)
}

func TestWorker_NewWorkerPrivKeyNilShouldFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	eligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, eligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilPrivateKey)
}

func TestWorker_NewWorkerPubKeyNilFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	wligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, wligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilPublicKey)
}

func TestWorker_NewWorkerShardCoordinatorNilFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	wligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, wligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		nil,
		mock.ValidatorGroupSelectorMock{},
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilShardCoordinator)
}

func TestWorker_NewWorkerValidatorGroupSelectorNilFail(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	wligibleList := createEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, wligibleList, consensusGroupSize, 0)
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

	wrk, err := bn.NewWorker(
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
		mock.ShardCoordinatorMock{},
		nil,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilValidatorGroupSelector)
}

func TestWorker_NewWorkerShouldWork(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := createEligibleList(consensusGroupSize)

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
		mock.ShardCoordinatorMock{},
		mock.ValidatorGroupSelectorMock{},
	)

	assert.NotNil(t, consensusWorker)
	assert.Nil(t, err)
}
