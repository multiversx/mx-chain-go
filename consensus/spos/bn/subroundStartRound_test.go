package bn_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func initSubroundStartRoundWithContainer(container spos.ConsensusCoreHandler) bn.SubroundStartRound {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	srStartRound, _ := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	return srStartRound
}

func initSubroundStartRound() bn.SubroundStartRound {
	container := mock.InitConsensusCore()
	return initSubroundStartRoundWithContainer(container)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srStartRound, err := bn.NewSubroundStartRound(
		nil,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetBlockchain(nil)
	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBootstraperShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetBootStrapper(nil)
	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBlootstraper, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	sr.ConsensusState = nil
	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetMultiSigner(nil)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetRounder(nil)
	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetSyncTimer(nil)
	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetValidatorGroupSelector(nil)
	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilValidatorGroupSelector, err)
}

func TestSubroundStartRound_NewSubroundStartRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.NotNil(t, srStartRound)
	assert.Nil(t, err)
}

func TestSubroundStartRound_DoStartRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	srStartRound, _ := bn.NewSubroundStartRound(
		sr,
		extend,
	)

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

	sr.SetStatus(bn.SrStartRound, spos.SsFinished)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenInitCurrentRoundReturnTrue(t *testing.T) {
	t.Parallel()

	boostraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(boostraperMock)

	sr := *initSubroundStartRoundWithContainer(container)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnFalseWhenInitCurrentRoundReturnFalse(t *testing.T) {
	t.Parallel()

	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return true
	}}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstraperMock)
	container.SetRounder(initRounderMock())

	sr := *initSubroundStartRoundWithContainer(container)

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenShouldSyncReturnTrue(t *testing.T) {
	t.Parallel()

	bootstraperMock := &mock.BootstraperMock{}

	bootstraperMock.ShouldSyncCalled = func() bool {
		return true
	}
	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstraperMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGenerateNextConsensusGroupErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	err := errors.New("error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte) ([]consensus.Validator, error) {
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

	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte) ([]consensus.Validator, error) {
		return make([]consensus.Validator, 0), nil
	}

	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	consensusState.SetSelfPubKey(consensusState.SelfPubKey() + "X")
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	srStartRound, _ := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
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

	bootstraperMock := &mock.BootstraperMock{}

	bootstraperMock.ShouldSyncCalled = func() bool {
		return false
	}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstraperMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_GenerateNextConsensusGroupShouldReturnErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	err := errors.New("error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte) ([]consensus.Validator, error) {
		return nil, err
	}
	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	err2 := srStartRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, err, err2)
}
