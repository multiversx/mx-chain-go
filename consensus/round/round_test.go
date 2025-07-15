package round_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/round"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"

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
		EnableEpochsHandler:       &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler:       &testscommon.EnableRoundsHandlerStub{},
	}
}

func TestRound_NewRoundShouldErrNilSyncTimer(t *testing.T) {
	t.Parallel()

	args := createDefaultRoundArgs()
	args.SyncTimer = nil
	rnd, err := round.NewRound(args)

	assert.Nil(t, rnd)
	assert.Equal(t, round.ErrNilSyncTimer, err)
}

func TestRound_NewRoundShouldWork(t *testing.T) {
	t.Parallel()

	args := createDefaultRoundArgs()
	rnd, err := round.NewRound(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(rnd))
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
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag != common.SupernovaFlag
			},
		}
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledCalled: func(flag common.EnableRoundFlag) bool {
				return flag != common.SupernovaRoundFlag
			},
		}

		args.GenesisTimeStamp = genesisTime

		rnd, _ := round.NewRound(args)
		oldIndex := rnd.Index()
		rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))
		newIndex := rnd.Index()

		assert.Equal(t, oldIndex, newIndex-1)
	})

	t.Run("after supernova", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SupernovaFlag
			},
		}
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

		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SupernovaFlag
			},
		}
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

	genesisTime := time.Now()

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime

	rnd, _ := round.NewRound(args)
	timeDuration := rnd.TimeDuration()

	assert.Equal(t, roundTimeDuration, timeDuration)
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
