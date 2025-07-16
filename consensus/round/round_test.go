package round_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/round"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"

	"github.com/stretchr/testify/assert"
)

const roundTimeDuration = 10 * time.Millisecond

func createDefaultRoundArgs() round.ArgsRound {
	genesisTime := time.Now()
	return round.ArgsRound{
		GenesisTimeStamp:          genesisTime,
		SupernovaGenesisTimeStamp: genesisTime,
		CurrentTimeStamp:          genesisTime,
		RoundTimeDuration:         roundTimeDuration,
		SupernovaTimeDuration:     roundTimeDuration,
		SyncTimer:                 &consensusMocks.SyncTimerMock{},
		StartRound:                0,
		SupernovaStartRound:       0,
		EnableRoundsHandler:       &testscommon.EnableRoundsHandlerStub{},
	}
}

func TestRound_NewRound(t *testing.T) {
	t.Parallel()

	t.Run("nil sync timer", func(t *testing.T) {
		t.Parallel()

		args := createDefaultRoundArgs()
		args.SyncTimer = nil
		rnd, err := round.NewRound(args)

		assert.Nil(t, rnd)
		assert.Equal(t, round.ErrNilSyncTimer, err)
	})

	t.Run("nil enable rounds handler", func(t *testing.T) {
		t.Parallel()

		args := createDefaultRoundArgs()
		args.EnableRoundsHandler = nil
		rnd, err := round.NewRound(args)

		assert.Nil(t, rnd)
		assert.Equal(t, errors.ErrNilEnableRoundsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultRoundArgs()
		rnd, err := round.NewRound(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(rnd))
	})
}

func TestRound_UpdateRoundShouldNotChangeAnything(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime

	rnd, _ := round.NewRound(args)
	oldIndex := rnd.Index()
	oldTimeStamp := rnd.TimeStamp()

	rnd.UpdateRound(genesisTime, genesisTime)

	newIndex := rnd.Index()
	newTimeStamp := rnd.TimeStamp()

	assert.Equal(t, oldIndex, newIndex)
	assert.Equal(t, oldTimeStamp, newTimeStamp)
}

func TestRound_UpdateRoundShouldAdvanceOneRound(t *testing.T) {
	t.Parallel()

	t.Run("before supernova", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledCalled: func(flag common.EnableRoundFlag) bool {
				return flag != common.SupernovaRoundFlag
			},
		}

		args.GenesisTimeStamp = genesisTime
		args.SupernovaGenesisTimeStamp = genesisTime.Add(10 * roundTimeDuration)

		rnd, _ := round.NewRound(args)
		oldIndex := rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))
		newIndex := rnd.Index()

		assert.Equal(t, oldIndex, newIndex-1)
	})

	t.Run("after supernova, with flag activated", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledCalled: func(flag common.EnableRoundFlag) bool {
				return flag == common.SupernovaRoundFlag
			},
		}

		args.SupernovaGenesisTimeStamp = genesisTime

		rnd, _ := round.NewRound(args)
		oldIndex := rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))
		newIndex := rnd.Index()

		assert.Equal(t, oldIndex, newIndex-1)
	})

	t.Run("after supernova, without flag activated, but after supernova genesis time", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		roundDuration := 10 * time.Millisecond

		supernovaRoundDuration := 5 * time.Millisecond
		supernovaStartRond := int64(5)
		supernovaGenesisTime := genesisTime.Add(time.Duration(supernovaStartRond) * roundDuration)

		args := createDefaultRoundArgs()
		args.RoundTimeDuration = roundDuration
		args.SupernovaTimeDuration = supernovaRoundDuration
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}

		args.SupernovaStartRound = supernovaStartRond
		args.GenesisTimeStamp = genesisTime

		args.SupernovaGenesisTimeStamp = genesisTime.Add(5 * roundDuration)

		rnd, _ := round.NewRound(args)

		rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))
		rnd.UpdateRound(genesisTime, genesisTime.Add(2*roundTimeDuration))
		rnd.UpdateRound(genesisTime, genesisTime.Add(3*roundTimeDuration))
		rnd.UpdateRound(genesisTime, genesisTime.Add(4*roundTimeDuration))

		index0 := rnd.Index()
		timestamp0 := rnd.TimeStamp()

		rnd.UpdateRound(genesisTime, supernovaGenesisTime)

		index1 := rnd.Index()
		timestamp1 := rnd.TimeStamp()

		diffTime := timestamp1.Sub(timestamp0)

		assert.Equal(t, index0, index1-1)
		assert.Equal(t, int64(5), index1)
		assert.Equal(t, roundDuration, diffTime)

		rnd.UpdateRound(genesisTime, supernovaGenesisTime.Add(supernovaRoundDuration))

		index2 := rnd.Index()
		timestamp2 := rnd.TimeStamp()

		diffTime2 := timestamp2.Sub(timestamp1)

		assert.Equal(t, index1, index2-1)
		assert.Equal(t, int64(6), index2)
		assert.Equal(t, supernovaRoundDuration, diffTime2)
	})
}

func TestRound_IndexShouldReturnFirstIndex(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime

	rnd, _ := round.NewRound(args)
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration/2))
	index := rnd.Index()

	assert.Equal(t, int64(0), index)
}

func TestRound_TimeStampShouldReturnTimeStampOfTheNextRound(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime
	args.SupernovaGenesisTimeStamp = genesisTime.Add(10 * roundTimeDuration)

	rnd, _ := round.NewRound(args)
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration+roundTimeDuration/2))
	timeStamp := rnd.TimeStamp()

	assert.Equal(t, genesisTime.Add(roundTimeDuration), timeStamp)
}

func TestRound_UpdateRoundWithTimeDurationChange(t *testing.T) {
	t.Parallel()

	t.Run("with transition to supernova in epoch", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.GenesisTimeStamp = genesisTime
		args.SupernovaGenesisTimeStamp = genesisTime.Add(10 * roundTimeDuration)

		rnd, _ := round.NewRound(args)

		oldIndex := rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))

		newIndex := rnd.Index()
		assert.Equal(t, oldIndex, newIndex-1)
		assert.Equal(t, int64(1), newIndex)

		oldIndex = rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(2*roundTimeDuration))

		newIndex = rnd.Index()
		assert.Equal(t, oldIndex, newIndex-1)
		assert.Equal(t, int64(2), newIndex)

		oldIndex = rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(3*roundTimeDuration))

		newIndex = rnd.Index()
		assert.Equal(t, oldIndex, newIndex-1)
		assert.Equal(t, int64(3), newIndex)
	})

	t.Run("with transition to supernova in round", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		flag := &atomic.Flag{}

		args := createDefaultRoundArgs()

		args.GenesisTimeStamp = genesisTime
		supernovaGenesisTime := genesisTime.Add(2 * roundTimeDuration)
		args.SupernovaGenesisTimeStamp = supernovaGenesisTime

		args.RoundTimeDuration = 10 * time.Millisecond
		newRoundTimeDuration := 5 * time.Millisecond
		args.SupernovaTimeDuration = newRoundTimeDuration

		args.StartRound = 0
		args.SupernovaStartRound = 2

		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledCalled: func(_ common.EnableRoundFlag) bool {
				return flag.IsSet()
			},
		}
		rnd, _ := round.NewRound(args)

		oldIndex := rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))

		newIndex := rnd.Index()
		assert.Equal(t, oldIndex, newIndex-1)
		assert.Equal(t, int64(1), newIndex)

		oldIndex = rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(2*roundTimeDuration))

		newIndex = rnd.Index()
		assert.Equal(t, oldIndex, newIndex-1)
		assert.Equal(t, int64(2), newIndex)

		flag.SetValue(true)

		oldIndex = rnd.Index()
		rnd.UpdateRound(supernovaGenesisTime, supernovaGenesisTime.Add(newRoundTimeDuration))

		newIndex = rnd.Index()
		assert.Equal(t, oldIndex, newIndex-1)
		assert.Equal(t, int64(3), newIndex)
	})
}

func TestRound_TimeDurationShouldReturnTheDurationOfOneRound(t *testing.T) {
	t.Parallel()

	t.Run("before supernova", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.GenesisTimeStamp = genesisTime

		rnd, _ := round.NewRound(args)
		timeDuration := rnd.TimeDuration()

		assert.Equal(t, roundTimeDuration, timeDuration)
	})

	t.Run("after supernova", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return flag == common.SupernovaRoundFlag
			},
		}

		args.SupernovaGenesisTimeStamp = genesisTime
		args.SupernovaTimeDuration = roundTimeDuration

		rnd, _ := round.NewRound(args)
		timeDuration := rnd.TimeDuration()

		assert.Equal(t, roundTimeDuration, timeDuration)
	})
}

func TestRound_RemainingTimeInCurrentRoundShouldReturnPositiveValue(t *testing.T) {
	t.Parallel()

	genesisTime := time.Unix(0, 0)

	syncTimerMock := &consensusMocks.SyncTimerMock{}

	timeElapsed := int64(roundTimeDuration - 1)

	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime
	args.CurrentTimeStamp = genesisTime
	args.SyncTimer = syncTimerMock

	rnd, _ := round.NewRound(args)

	remainingTime := rnd.RemainingTime(rnd.TimeStamp(), roundTimeDuration)

	assert.Equal(t, time.Duration(int64(rnd.TimeDuration())-timeElapsed), remainingTime)
	assert.True(t, remainingTime > 0)
}

func TestRound_RemainingTimeInCurrentRoundShouldReturnNegativeValue(t *testing.T) {
	t.Parallel()

	genesisTime := time.Unix(0, 0)

	syncTimerMock := &consensusMocks.SyncTimerMock{}

	timeElapsed := int64(roundTimeDuration + 1)

	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime
	args.CurrentTimeStamp = genesisTime
	args.SyncTimer = syncTimerMock

	rnd, _ := round.NewRound(args)

	remainingTime := rnd.RemainingTime(rnd.TimeStamp(), roundTimeDuration)

	assert.Equal(t, time.Duration(int64(rnd.TimeDuration())-timeElapsed), remainingTime)
	assert.True(t, remainingTime < 0)
}

func TestRound_RevertOneRound(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()

	syncTimerMock := &consensusMocks.SyncTimerMock{}

	startRound := int64(10)

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime
	args.SyncTimer = syncTimerMock
	args.StartRound = startRound

	rnd, _ := round.NewRound(args)
	index := rnd.Index()
	require.Equal(t, startRound, index)

	rnd.RevertOneRound()
	index = rnd.Index()
	require.Equal(t, startRound-1, index)
}

func TestRound_BeforeGenesis(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()

	syncTimerMock := &consensusMocks.SyncTimerMock{}

	startRound := int64(-1)
	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime
	args.SyncTimer = syncTimerMock
	args.StartRound = startRound

	rnd, _ := round.NewRound(args)
	require.True(t, rnd.BeforeGenesis())

	time.Sleep(roundTimeDuration * 2)
	currentTime := time.Now()

	rnd.UpdateRound(genesisTime, currentTime)
	require.False(t, rnd.BeforeGenesis())
}
