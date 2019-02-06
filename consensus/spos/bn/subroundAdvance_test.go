package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestWorker_DoAdvanceJobShouldReturnFalse(t *testing.T) {
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

	worker.SPoS.SetStatus(bn.SrEndRound, spos.SsFinished)

	assert.False(t, worker.DoAdvanceJob())
}

func TestWorker_DoAdvanceJobShouldReturnTrue(t *testing.T) {
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

	assert.True(t, worker.DoAdvanceJob())
}
