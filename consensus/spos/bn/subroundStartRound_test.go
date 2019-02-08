package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func initSubroundStartRound() bn.SubroundStartRound {
	blockChain := blockchain.BlockChain{}
	//blockProcessorMock := initBlockProcessorMock()
	//bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
	//	return false
	//}}

	consensusState := initConsensusState()
	//hasherMock := mock.HasherMock{}
	//marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	//shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
	)

	srStartRound, _ := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	return srStartRound
}

func TestWorker_StartRound(t *testing.T) {
	sr := *initSubroundStartRound()

	r := sr.DoStartRoundJob()
	assert.True(t, r)
}

func TestWorker_CheckStartRoundConsensus(t *testing.T) {
	sr := *initSubroundStartRound()

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}
