package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func TestWorker_DoEndRoundJobNotFinished(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].Header = &block.Header{}

	r := cnWorkers[0].DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
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

	multisigner.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return nil, crypto.ErrNilHasher
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

	worker.SendMessage = SendMessage
	worker.BroadcastBlockBody = BroadcastMessage
	worker.BroadcastHeader = BroadcastMessage

	worker.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, sPoS, worker)
	worker.Header = &block.Header{}

	r := worker.DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
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

	worker := initSposWorker(sPoS)

	blProcMock := initMockBlockProcessor()

	blProcMock.CommitBlockCalled = func(
		blockChain *blockchain.BlockChain,
		header *block.Header,
		block *block.TxBlockBody,
	) error {
		return blockchain.ErrHeaderUnitNil
	}

	worker.BlockProcessor = blProcMock

	worker.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, sPoS, worker)
	worker.Header = &block.Header{}

	r := worker.DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrRemBlockTxOK(t *testing.T) {
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

	worker := initSposWorker(sPoS)

	blProcMock := initMockBlockProcessor()

	blProcMock.RemoveBlockTxsFromPoolCalled = func(body *block.TxBlockBody) error {
		return process.ErrNilBlockBodyPool
	}

	worker.BlockProcessor = blProcMock

	worker.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, sPoS, worker)
	worker.Header = &block.Header{}

	r := worker.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobErrBroadcastTxBlockBodyOK(t *testing.T) {
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

	worker := initSposWorker(sPoS)

	worker.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	worker.BroadcastBlockBody = nil

	GenerateSubRoundHandlers(roundDuration, sPoS, worker)
	worker.Header = &block.Header{}

	r := worker.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobErrBroadcastHeaderOK(t *testing.T) {
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

	worker := initSposWorker(sPoS)

	worker.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	worker.BroadcastHeader = nil

	GenerateSubRoundHandlers(roundDuration, sPoS, worker)
	worker.Header = &block.Header{}

	r := worker.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobAllOK(t *testing.T) {
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

	worker := initSposWorker(sPoS)

	worker.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	worker.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	GenerateSubRoundHandlers(roundDuration, sPoS, worker)
	worker.Header = &block.Header{}

	r := worker.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_CheckEndRoundConsensus(t *testing.T) {
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

	ok := worker.CheckEndRoundConsensus()
	assert.True(t, ok)
}

func TestWorker_ExtendEndRound(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].ExtendEndRound()
}

func TestWorker_CheckSignaturesValidityShouldErrNilSignature(t *testing.T) {
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

	err := worker.CheckSignaturesValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestWorker_CheckSignaturesValidityShouldErrInvalidIndex(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 22
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

	worker.SPoS.SetJobDone(consensusGroup[0], bn.SrSignature, true)

	err := worker.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrInvalidIndex, err)
}

func TestWorker_CheckSignaturesValidityShouldRetunNil(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 22
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

	worker.SPoS.SetJobDone(consensusGroup[0], bn.SrSignature, true)

	err := worker.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}
