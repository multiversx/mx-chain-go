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

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	wrk, _ := bn.NewWorker(
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

	bnf, _ := bn.NewbnFactory(
		wrk,
	)

	bnf.GenerateSubrounds()
	assert.Equal(t, 8, len(wrk.SPoS.Chr.SubroundHandlers()))
}

func TestFactory_NewbnFactoryNilShouldFail(t *testing.T) {

	bnf, err := bn.NewbnFactory(
		nil,
	)

	assert.Nil(t, bnf)
	assert.Equal(t, err, spos.ErrNilWorker)
}

func TestFactory_NewbnFactoryShouldWork(t *testing.T) {
	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(genesisTime, roundDuration, consensusGroup, consensusGroupSize, 0)
	blockChain := &blockchain.BlockChain{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	blkProc := &mock.BlockProcessorMock{}
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}
	multisig := mock.NewMultiSigner()
	keyGen := &mock.KeyGenMock{}
	privKey := &mock.PrivateKeyMock{}
	pubKey := &mock.PublicKeyMock{}

	wrk, _ := bn.NewWorker(
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

	bnf, _ := bn.NewbnFactory(
		wrk,
	)

	assert.NotNil(t, bnf)
}
