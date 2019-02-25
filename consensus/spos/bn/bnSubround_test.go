package bn_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

var jobSuccess = true

func Job() bool {
	fmt.Printf("do job with success: %v\n", jobSuccess)
	return jobSuccess
}

var mutCheck sync.RWMutex
var checkSuccess = true

func Check() bool {
	mutCheck.RLock()
	defer mutCheck.RUnlock()

	fmt.Printf("do check consensus with success: %v\n", checkSuccess)
	return checkSuccess
}

func TestSubround_NewSubroundNilChannelShouldFail(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	assert.Nil(t, err)
	assert.NotNil(t, sr)
}

func TestSubround_DoWorkShouldReturnFalseWhenJobFunctionIsNotSet(t *testing.T) {
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

	checkSuccess = true

	sr.SetJobFunction(nil)
	sr.SetCheckFunction(Check)

	maxTime := time.Now().Add(100 * time.Millisecond)
	haveTime := func() time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(haveTime)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenCheckFunctionIsNotSet(t *testing.T) {
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

	jobSuccess = true

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(nil)

	maxTime := time.Now().Add(100 * time.Millisecond)
	haveTime := func() time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(haveTime)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenConsensusIsNotDone(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	maxTime := time.Now().Add(100 * time.Millisecond)
	haveTime := func() time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(haveTime)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobAndConsensusAreDone(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = true

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	maxTime := time.Now().Add(100 * time.Millisecond)
	haveTime := func() time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(haveTime)
	assert.True(t, r)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobIsDoneAndConsensusIsDoneAfterAWhile(t *testing.T) {
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

	jobSuccess = true

	mutCheck.Lock()
	checkSuccess = false
	mutCheck.Unlock()

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	maxTime := time.Now().Add(100 * time.Millisecond)
	haveTime := func() time.Duration {
		return maxTime.Sub(time.Now())
	}

	go func() {
		time.Sleep(50 * time.Millisecond)

		mutCheck.Lock()
		checkSuccess = true
		mutCheck.Unlock()

		ch <- true
	}()

	r := sr.DoWork(haveTime)

	assert.True(t, r)
}

func TestSubround_Previous(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	assert.Equal(t, int(bn.SrStartRound), sr.Previous())
}

func TestSubround_Current(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	assert.Equal(t, int(bn.SrBlock), sr.Current())
}

func TestSubround_Next(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	assert.Equal(t, int(bn.SrCommitmentHash), sr.Next())
}

func TestSubround_StartTime(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.StartTime())
}

func TestSubround_EndTime(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.EndTime())
}

func TestSubround_Name(t *testing.T) {
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

	jobSuccess = true
	checkSuccess = false

	sr.SetJobFunction(Job)
	sr.SetCheckFunction(Check)

	assert.Equal(t, "(BLOCK)", sr.Name())
}
