package bn_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func TestSubround_NewSubroundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitContainer()
	ch := make(chan bool, 1)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		nil,
		ch,
		container,
	)

	assert.Equal(t, spos.ErrNilConsensusState, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundNilChannelShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitContainer()

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		nil,
		container,
	)

	assert.Equal(t, spos.ErrNilChannel, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundNilContainerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		nil,
	)

	assert.Equal(t, spos.ErrNilConsensusDataContainer, err)
	assert.Nil(t, sr)
}

func TestSubround_NilContainerBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetBlockchain(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubround_NilContainerBlockprocessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetBlockProcessor(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubround_NilContainerBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetBootStrapper(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlootstraper, err)
}

func TestSubround_NilContainerChronologyShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetChronology(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestSubround_NilContainerHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetHasher(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubround_NilContainerMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetMarshalizer(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestSubround_NilContainerMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetMultiSigner(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubround_NilContainerRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetRounder(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubround_NilContainerShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetShardCoordinator(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestSubround_NilContainerSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetSyncTimer(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubround_NilContainerValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetValidatorGroupSelector(nil)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilValidatorGroupSelector, err)
}

func TestSubround_NewSubroundShouldWork(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, err)

	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.NotNil(t, sr)
}

func TestSubround_DoWorkShouldReturnFalseWhenJobFunctionIsNotSet(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(nil)
	sr.SetCheckFunction(func() bool {
		return true
	})

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)

	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenCheckFunctionIsNotSet(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(nil)

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenConsensusIsNotDone(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobAndConsensusAreDone(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		-1,
		bn.SrStartRound,
		bn.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return true
	})

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)
	assert.True(t, r)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobIsDoneAndConsensusIsDoneAfterAWhile(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		-1,
		bn.SrStartRound,
		bn.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	var mut sync.RWMutex
	mut.Lock()
	checkSuccess := false
	mut.Unlock()

	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		mut.RLock()
		defer mut.RUnlock()
		return checkSuccess
	})

	maxTime := time.Now().Add(2000 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	go func() {
		time.Sleep(1000 * time.Millisecond)

		mut.Lock()
		checkSuccess = true
		mut.Unlock()

		ch <- true
	}()

	r := sr.DoWork(rounderMock)

	assert.True(t, r)
}

func TestSubround_Previous(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Equal(t, int(bn.SrStartRound), sr.Previous())
}

func TestSubround_Current(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Equal(t, int(bn.SrBlock), sr.Current())
}

func TestSubround_Next(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Equal(t, int(bn.SrCommitmentHash), sr.Next())
}

func TestSubround_StartTime(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetRounder(initRounderMock())
	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(25*roundTimeDuration/100),
		int64(40*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.StartTime())
}

func TestSubround_EndTime(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()
	container.SetRounder(initRounderMock())
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})

	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.EndTime())
}

func TestSubround_Name(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitContainer()

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Equal(t, "(BLOCK)", sr.Name())
}
