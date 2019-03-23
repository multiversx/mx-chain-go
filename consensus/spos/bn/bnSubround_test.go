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

func TestSubround_NewSubroundNilChannelShouldFail(t *testing.T) {
	t.Parallel()

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		nil,
	)

	assert.Equal(t, spos.ErrNilChannel, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundShouldWork(t *testing.T) {
	t.Parallel()

	ch := make(chan bool, 1)

	sr, err := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
	)

	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Nil(t, err)
	assert.NotNil(t, sr)
}

func TestSubround_DoWorkShouldReturnFalseWhenJobFunctionIsNotSet(t *testing.T) {
	t.Parallel()

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(-1),
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		bn.SrStartRound,
		bn.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		-1,
		bn.SrStartRound,
		bn.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(25*roundTimeDuration/100),
		int64(40*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		ch,
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

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		ch,
	)

	sr.SetJobFunction(func() bool {
		return true
	})
	sr.SetCheckFunction(func() bool {
		return false
	})

	assert.Equal(t, "(BLOCK)", sr.Name())
}
