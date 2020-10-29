package bls_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func defaultSubroundStartRoundFromSubround(sr *spos.Subround) (bls.SubroundStartRound, error) {
	startRound, err := bls.NewSubroundStartRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		executeStoredMessages,
		resetConsensusMessages,
	)

	return startRound, err
}

func defaultWithoutErrorSubroundStartRoundFromSubround(sr *spos.Subround) bls.SubroundStartRound {
	startRound, _ := bls.NewSubroundStartRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		executeStoredMessages,
		resetConsensusMessages,
	)

	return startRound
}

func defaultSubround(
	consensusState *spos.ConsensusState,
	ch chan bool,
	container spos.ConsensusCoreHandler,
) (*spos.Subround, error) {

	return spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&mock.AppStatusHandlerStub{},
	)
}

func initSubroundStartRoundWithContainer(container spos.ConsensusCoreHandler) bls.SubroundStartRound {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubround(consensusState, ch, container)
	srStartRound, _ := bls.NewSubroundStartRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		executeStoredMessages,
		resetConsensusMessages,
	)

	return srStartRound
}

func initSubroundStartRound() bls.SubroundStartRound {
	container := mock.InitConsensusCore()
	return initSubroundStartRoundWithContainer(container)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srStartRound, err := bls.NewSubroundStartRound(
		nil,
		extend,
		bls.ProcessingThresholdPercent,
		executeStoredMessages,
		resetConsensusMessages,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetBlockchain(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetBootStrapper(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	sr.ConsensusState = nil
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetMultiSigner(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetRounder(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetSyncTimer(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetValidatorGroupSelector(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestSubroundStartRound_NewSubroundStartRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.NotNil(t, srStartRound)
	assert.Nil(t, err)
}

func TestSubroundStartRound_DoStartRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	srStartRound := *defaultWithoutErrorSubroundStartRoundFromSubround(sr)

	r := srStartRound.DoStartRoundJob()
	assert.True(t, r)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundStartRound()

	sr.RoundCanceled = true

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundStartRound()

	sr.SetStatus(bls.SrStartRound, spos.SsFinished)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenInitCurrentRoundReturnTrue(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperMock{GetNodeStateCalled: func() core.NodeState {
		return core.NsSynchronized
	}}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	sr := *initSubroundStartRoundWithContainer(container)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnFalseWhenInitCurrentRoundReturnFalse(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperMock{GetNodeStateCalled: func() core.NodeState {
		return core.NsNotSynchronized
	}}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)
	container.SetRounder(initRounderMock())

	sr := *initSubroundStartRoundWithContainer(container)

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGetNodeStateNotReturnSynchronized(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperMock{}

	bootstrapperMock.GetNodeStateCalled = func() core.NodeState {
		return core.NsNotSynchronized
	}
	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGenerateNextConsensusGroupErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &mock.NodesCoordinatorMock{}
	err := errors.New("error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte, round uint64, shardId uint32, epoch uint32) ([]sharding.Validator, error) {
		return nil, err
	}
	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGetLeaderErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &mock.NodesCoordinatorMock{}
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(
		bytes []byte,
		round uint64,
		shardId uint32,
		epoch uint32,
	) ([]sharding.Validator, error) {
		return make([]sharding.Validator, 0), nil
	}

	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrueWhenIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	consensusState.SetSelfPubKey(consensusState.SelfPubKey() + "X")
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	srStartRound := *defaultWithoutErrorSubroundStartRoundFromSubround(sr)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenCreateErr(t *testing.T) {
	t.Parallel()

	multiSignerMock := mock.InitMultiSignerMock()
	err := errors.New("error")
	multiSignerMock.ResetCalled = func(pubKeys []string, index uint16) error {
		return err
	}

	container := mock.InitConsensusCore()
	container.SetMultiSigner(multiSignerMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenTimeIsOut(t *testing.T) {
	t.Parallel()

	rounderMock := initRounderMock()

	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return time.Duration(-1)
	}

	container := mock.InitConsensusCore()
	container.SetRounder(rounderMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperMock{}

	bootstrapperMock.GetNodeStateCalled = func() core.NodeState {
		return core.NsSynchronized
	}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_GenerateNextConsensusGroupShouldReturnErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &mock.NodesCoordinatorMock{}

	err := errors.New("error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(
		bytes []byte,
		round uint64,
		shardId uint32,
		epoch uint32,
	) ([]sharding.Validator, error) {
		return nil, err
	}
	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	err2 := srStartRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, err, err2)
}
