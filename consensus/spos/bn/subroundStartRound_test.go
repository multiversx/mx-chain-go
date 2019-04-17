package bn_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func initSubroundStartRoundWithContainer(container spos.ConsensusDataContainerInterface) bn.SubroundStartRound {
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
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
	container := mock.InitContainer()
	return initSubroundStartRoundWithContainer(container)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srStartRound, err := bn.NewSubroundStartRound(
		nil,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilSubround)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	container.SetBlockchain(nil)

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilBlockChain)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBootstraperShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	container.SetBootStrapper(nil)

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilBlootstraper)
}

func TestSubroundStartRound_NewSubroundStartRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	container.SetConsensusState(nil)

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestSubroundStartRound_NewSubroundStartRoundNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	container.SetMultiSigner(nil)

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestSubroundStartRound_NewSubroundStartRoundNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	container.SetRounder(nil)

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	container.SetSyncTimer(nil)

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestSubroundStartRound_NewSubroundStartRoundNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	container.SetValidatorGroupSelector(nil)

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
		container,
	)

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilValidatorGroupSelector)
}

func TestSubroundStartRound_NewSubroundStartRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
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

	container := mock.InitContainer()

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
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

	sr.ConsensusState().RoundCanceled = true

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundStartRound()

	sr.ConsensusState().SetStatus(bn.SrStartRound, spos.SsFinished)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenInitCurrentRoundReturnTrue(t *testing.T) {
	t.Parallel()

	boostraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	container := mock.InitContainer()
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

	container := mock.InitContainer()
	container.SetBootStrapper(bootstraperMock)

	sr := *initSubroundStartRound()

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenShouldSyncReturnTrue(t *testing.T) {
	t.Parallel()

	bootstraperMock := &mock.BootstraperMock{}

	bootstraperMock.ShouldSyncCalled = func() bool {
		return true
	}
	container := mock.InitContainer()
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
	container := mock.InitContainer()
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

	container := mock.InitContainer()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	consensusState.SetSelfPubKey(consensusState.SelfPubKey() + "X")

	container := mock.InitContainer()
	container.SetConsensusState(consensusState)

	srStartRound := *initSubroundStartRoundWithContainer(container)

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

	container := mock.InitContainer()
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

	container := mock.InitContainer()
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

	container := mock.InitContainer()
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
	container := mock.InitContainer()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	err2 := srStartRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, err, err2)
}
