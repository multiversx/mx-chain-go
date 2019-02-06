package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWorker_SendSignature(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.SetSelfPubKey(cnWorkers[0].SPoS.ConsensusGroup()[0])

	r := cnWorkers[0].DoSignatureJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	cnWorkers[0].SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	r = cnWorkers[0].DoSignatureJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrSignature, spos.SsNotFinished)
	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrSignature, true)

	r = cnWorkers[0].DoSignatureJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrSignature, false)

	r = cnWorkers[0].DoSignatureJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBitmap, true)
	cnWorkers[0].SPoS.Data = nil

	r = cnWorkers[0].DoSignatureJob()
	assert.False(t, r)

	dta := []byte("X")
	cnWorkers[0].SPoS.Data = dta

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrCommitment, true)
	r = cnWorkers[0].DoSignatureJob()
	assert.True(t, r)
}

func TestWorker_CheckCommitmentsValidityShouldErrNilCommitmet(t *testing.T) {
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
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
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

	err := worker.CheckCommitmentsValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilCommitment, err)
}

func TestWorker_CheckCommitmentsValidityShouldErrInvalidIndex(t *testing.T) {
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
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	multisigner.Reset(nil, 0)

	worker, _ := bn.NewWorker(
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

	worker.SPoS.SetJobDone(consensusGroup[0], bn.SrCommitment, true)

	err := worker.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrInvalidIndex, err)
}

func TestWorker_CheckCommitmentsValidityShouldErrOnCommitmentHash(t *testing.T) {
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
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	err := errors.New("error commitment hash")
	multisigner.CommitmentHashMock = func(uint16) ([]byte, error) {
		return nil, err
	}

	worker, _ := bn.NewWorker(
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

	worker.SPoS.SetJobDone(consensusGroup[0], bn.SrCommitment, true)

	err2 := worker.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, err, err2)
}

func TestWorker_CheckCommitmentsValidityShouldErrCommitmentHashDoesNotMatch(t *testing.T) {
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
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	multisigner.CommitmentHashMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	worker, _ := bn.NewWorker(
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

	worker.SPoS.SetJobDone(consensusGroup[0], bn.SrCommitment, true)

	err := worker.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, spos.ErrCommitmentHashDoesNotMatch, err)
}

func TestWorker_CheckCommitmentsValidityShouldReturnNil(t *testing.T) {
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
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	multisigner.CommitmentMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	multisigner.CommitmentHashMock = func(uint16) ([]byte, error) {
		return mock.HasherMock{}.Compute(string([]byte("X"))), nil
	}

	worker, _ := bn.NewWorker(
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

	worker.SPoS.SetJobDone(consensusGroup[0], bn.SrCommitment, true)

	err := worker.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}

func TestWorker_ReceivedSignature(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		cnWorkers[0].SPoS.Data,
		nil,
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	r := cnWorkers[0].ReceivedSignature(cnsDta)
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrSignature, spos.SsNotFinished)

	r = cnWorkers[0].ReceivedSignature(cnsDta)
	assert.False(t, r)

	cnWorkers[0].SPoS.RoundConsensus.SetJobDone(cnWorkers[0].SPoS.ConsensusGroup()[0], bn.SrBitmap, true)

	r = cnWorkers[0].ReceivedSignature(cnsDta)
	assert.True(t, r)

	isSignJobDone, _ := cnWorkers[0].SPoS.GetJobDone(cnWorkers[0].SPoS.ConsensusGroup()[0], bn.SrSignature)
	assert.True(t, isSignJobDone)
}

func TestWorker_CheckSignatureConsensus(t *testing.T) {
	sPoS := InitSposWorker()

	worker, _ := bn.NewWorker(
		sPoS,
		&blockchain.BlockChain{},
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BootstrapMock{ShouldSyncCalled: func() bool {
			return false
		}},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, sPoS, worker)

	sPoS.SetStatus(bn.SrSignature, spos.SsNotFinished)

	ok := worker.CheckSignatureConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrSignature))

	for i := 0; i < sPoS.Threshold(bn.SrBitmap); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	for i := 1; i < sPoS.Threshold(bn.SrSignature); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrSignature, true)
	}

	ok = worker.CheckSignatureConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrSignature))

	sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[0], bn.SrSignature, true)

	ok = worker.CheckSignatureConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrSignature))
}

func TestWorker_SignaturesCollected(t *testing.T) {
	sPoS := InitSposWorker()

	worker, _ := bn.NewWorker(
		sPoS,
		&blockchain.BlockChain{},
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BootstrapMock{ShouldSyncCalled: func() bool {
			return false
		}},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, sPoS, worker)

	for i := 0; i < len(worker.SPoS.ConsensusGroup()); i++ {
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := worker.SignaturesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrBitmap, true)
	worker.SPoS.SetJobDone("3", bn.SrBitmap, true)
	isJobDone, _ := worker.SPoS.GetJobDone("3", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = worker.SignaturesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("2", bn.SrSignature, true)
	isJobDone, _ = worker.SPoS.GetJobDone("2", bn.SrSignature)
	assert.True(t, isJobDone)

	ok = worker.SignaturesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("3", bn.SrSignature, true)
	ok = worker.SignaturesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrSignature, true)
	ok = worker.SignaturesCollected(2)
	assert.True(t, ok)
}

func TestWorker_ExtendSignature(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].ExtendSignature()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].SPoS.Status(bn.SrSignature))
}
