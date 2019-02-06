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

func TestWorker_SendCommitmentHash(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.SetSelfPubKey(cnWorkers[0].SPoS.ConsensusGroup()[0])

	r := cnWorkers[0].DoCommitmentHashJob()
	assert.True(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	cnWorkers[0].SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrCommitmentHash, true)

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrCommitmentHash, false)
	cnWorkers[0].SPoS.Data = nil

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.False(t, r)

	dta := []byte("X")
	cnWorkers[0].SPoS.Data = dta

	r = cnWorkers[0].DoCommitmentHashJob()
	assert.True(t, r)
}

func TestWorker_DoCommitmentHashJobErrCreateCommitmentShouldFail(t *testing.T) {
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

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, crypto.ErrNilHasher
	}

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

	cnWorker.SendMessage = SendMessage
	cnWorker.BroadcastBlockBody = BroadcastMessage
	cnWorker.BroadcastHeader = BroadcastMessage

	cnWorker.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	cnWorker.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	done := cnWorker.DoCommitmentHashJob()

	assert.False(t, done)
}

func TestWorker_ReceivedCommitmentHash(t *testing.T) {
	cnWorkers := initSposWorkers()
	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	dta := []byte("X")

	cnsDta := spos.NewConsensusData(
		dta,
		nil,
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[0]),
		nil,
		int(bn.MtCommitmentHash),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	for i := 0; i < cnWorkers[0].SPoS.Threshold(bn.SrCommitmentHash); i++ {
		cnWorkers[0].SPoS.RoundConsensus.SetJobDone(cnWorkers[0].SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	r := cnWorkers[0].ReceivedCommitmentHash(cnsDta)
	assert.False(t, r)

	for i := 0; i < cnWorkers[0].SPoS.Threshold(bn.SrCommitmentHash); i++ {
		cnWorkers[0].SPoS.RoundConsensus.SetJobDone(cnWorkers[0].SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
	}

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	r = cnWorkers[0].ReceivedCommitmentHash(cnsDta)
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	r = cnWorkers[0].ReceivedCommitmentHash(cnsDta)
	assert.True(t, r)
	isCommHashJobDone, _ := cnWorkers[0].SPoS.GetJobDone(cnWorkers[0].SPoS.ConsensusGroup()[0], bn.SrCommitmentHash)
	assert.True(t, isCommHashJobDone)
}

func TestWorker_CheckCommitmentHashConsensus(t *testing.T) {
	sPoS := InitSposWorker()

	sPoS.SetSelfPubKey(sPoS.ConsensusGroup()[0])

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

	sPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok := worker.CheckCommitmentHashConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < sPoS.Threshold(bn.SrCommitmentHash); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	ok = worker.CheckCommitmentHashConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrCommitmentHash))

	sPoS.RoundConsensus.SetSelfPubKey("2")

	sPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok = worker.CheckCommitmentHashConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < sPoS.Threshold(bn.SrBitmap); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	ok = worker.CheckCommitmentHashConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < sPoS.Threshold(bn.SrBitmap); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, false)
	}

	for i := 0; i < len(sPoS.RoundConsensus.ConsensusGroup()); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	sPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok = worker.CheckCommitmentHashConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrCommitmentHash))
}

func TestWorker_IsCommitmentHashReceived(t *testing.T) {
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

	ok := worker.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrCommitmentHash, true)
	isJobDone, _ := worker.SPoS.GetJobDone("1", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = worker.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("2", bn.SrCommitmentHash, true)
	ok = worker.IsCommitmentHashReceived(2)
	assert.True(t, ok)

	worker.SPoS.SetJobDone("3", bn.SrCommitmentHash, true)
	ok = worker.IsCommitmentHashReceived(2)
	assert.True(t, ok)
}

func TestWorker_CommitmentHashesCollected(t *testing.T) {
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

	ok := worker.CommitmentHashesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrBitmap, true)
	worker.SPoS.SetJobDone("3", bn.SrBitmap, true)
	isJobDone, _ := worker.SPoS.GetJobDone("3", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = worker.CommitmentHashesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("2", bn.SrCommitmentHash, true)
	isJobDone, _ = worker.SPoS.GetJobDone("2", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = worker.CommitmentHashesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("3", bn.SrCommitmentHash, true)
	ok = worker.CommitmentHashesCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrCommitmentHash, true)
	ok = worker.CommitmentHashesCollected(2)
	assert.True(t, ok)
}

func TestWorker_ExtendCommitmentHash(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].SPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < len(cnWorkers[0].SPoS.RoundConsensus.ConsensusGroup()); i++ {
		cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	cnWorkers[0].ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].SPoS.Status(bn.SrCommitmentHash))
}

func TestWorker_GenCommitmentHashShouldRetunErrOnCreateCommitment(t *testing.T) {
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

	err := errors.New("error create commitment")

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, err
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

	_, err2 := worker.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnIndexSelfConsensusGroup(t *testing.T) {
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

	sPoS.SetSelfPubKey("X")

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multisigner.AddCommitmentMock = func(uint16, []byte) error {
		return spos.ErrSelfNotFoundInConsensus
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

	_, err := worker.GenCommitmentHash()
	assert.Equal(t, spos.ErrSelfNotFoundInConsensus, err)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnAddCommitment(t *testing.T) {
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

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	err := errors.New("error add commitment")

	multisigner.AddCommitmentMock = func(uint16, []byte) error {
		return err
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

	_, err2 := worker.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnSetCommitmentSecret(t *testing.T) {
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

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multisigner.AddCommitmentMock = func(uint16, []byte) error {
		return nil
	}

	err := errors.New("error set commitment secret")

	multisigner.SetCommitmentSecretMock = func([]byte) error {
		return err
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

	_, err2 := worker.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnAddCommitmentHash(t *testing.T) {
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

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multisigner.AddCommitmentMock = func(uint16, []byte) error {
		return nil
	}

	multisigner.SetCommitmentSecretMock = func([]byte) error {
		return nil
	}

	err := errors.New("error add commitment hash")

	multisigner.AddCommitmentHashMock = func(uint16, []byte) error {
		return err
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

	_, err2 := worker.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunNil(t *testing.T) {
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

	multisigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multisigner.AddCommitmentMock = func(uint16, []byte) error {
		return nil
	}

	multisigner.SetCommitmentSecretMock = func([]byte) error {
		return nil
	}

	multisigner.AddCommitmentHashMock = func(uint16, []byte) error {
		return nil
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

	_, err := worker.GenCommitmentHash()
	assert.Equal(t, nil, err)
}
