package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func initSubroundStartRound() bn.SubroundStartRound {
	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	return srStartRound
}

func TestSubroundStartRound_NewSubroundStartRoundNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	srStartRound, err := bn.NewSubroundStartRound(
		nil,
		&blockChain,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilSubround)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		nil,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilBlockChain)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBootstraperShouldFail(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		nil,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilBlootstraper)
}

func TestSubroundStartRound_NewSubroundStartRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		nil,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestSubroundStartRound_NewSubroundStartRoundNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	rounderMock := initRounderMock()
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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		consensusState,
		nil,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestSubroundStartRound_NewSubroundStartRoundNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		nil,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		nil,
		validatorGroupSelector,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestSubroundStartRound_NewSubroundStartRoundNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		nil,
		extend,
	)

	assert.Nil(t, srStartRound)
	assert.Equal(t, err, spos.ErrNilValidatorGroupSelector)
}

func TestSubroundStartRound_NewSubroundStartRoundShouldWork(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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

	srStartRound, err := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	assert.NotNil(t, srStartRound)
	assert.Nil(t, err)
}

func TestSubroundStartRound_DoStartRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
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

	sr := *initSubroundStartRound()

	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	sr.SetBootsraper(bootstraperMock)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnFalseWhenInitCurrentRoundReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundStartRound()

	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return true
	}}

	sr.SetBootsraper(bootstraperMock)

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenShouldSyncReturnTrue(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	bootstraperMock.ShouldSyncCalled = func() bool {
		return true
	}

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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGenerateNextConsensusGroupErr(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	err := errors.New("Error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte) ([]consensus.Validator, error) {
		return nil, err
	}

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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGetLeaderErr(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte) ([]consensus.Validator, error) {
		return make([]consensus.Validator, 0), nil
	}

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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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

	consensusState.SetSelfPubKey(consensusState.SelfPubKey() + "X")

	srStartRound, _ := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenCreateErr(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
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

	err := errors.New("error")
	multiSignerMock.ResetCalled = func(pubKeys []string, index uint16) error {
		return err
	}

	srStartRound, _ := bn.NewSubroundStartRound(
		sr,
		&blockChain,
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenTimeIsOut(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	rounderMock.RemainingTimeInRoundCalled = func(safeThresholdPercent uint32) time.Duration {
		return time.Duration(-1)
	}

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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	bootstraperMock.ShouldSyncCalled = func() bool {
		return false
	}

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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_GenerateNextConsensusGroupShouldReturnErr(t *testing.T) {
	t.Parallel()

	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	err := errors.New("Error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte) ([]consensus.Validator, error) {
		return nil, err
	}

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
		bootstraperMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		validatorGroupSelector,
		extend,
	)

	err2 := srStartRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, err, err2)
}
