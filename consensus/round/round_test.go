package round_test

import (
	"sync"
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
	args.SupernovaGenesisTimeStamp = genesisTime.Add(10 * roundTimeDuration)

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
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
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

		roundDuration := 10 * time.Millisecond

		supernovaRoundDuration := 5 * time.Millisecond
		supernovaStartRond := int64(5)
		supernovaGenesisTime := genesisTime.Add(time.Duration(supernovaStartRond) * roundDuration)

		args := createDefaultRoundArgs()
		args.RoundTimeDuration = roundDuration
		args.SupernovaTimeDuration = supernovaRoundDuration
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return flag == common.SupernovaRoundFlag && round >= uint64(supernovaStartRond)
			},
		}

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
		assert.Equal(t, roundDuration, rnd.TimeDuration())

		rnd.UpdateRound(genesisTime, supernovaGenesisTime)

		index1 := rnd.Index()
		timestamp1 := rnd.TimeStamp()

		diffTime := timestamp1.Sub(timestamp0)

		assert.Equal(t, supernovaRoundDuration, rnd.TimeDuration())
		assert.Equal(t, index0, index1-1)
		assert.Equal(t, int64(5), index1)
		assert.Equal(t, roundDuration, diffTime)

		rnd.UpdateRound(genesisTime, supernovaGenesisTime.Add(supernovaRoundDuration))

		index2 := rnd.Index()
		timestamp2 := rnd.TimeStamp()

		diffTime2 := timestamp2.Sub(timestamp1)

		assert.Equal(t, supernovaRoundDuration, rnd.TimeDuration())
		assert.Equal(t, index1, index2-1)
		assert.Equal(t, int64(6), index2)
		assert.Equal(t, supernovaRoundDuration, diffTime2)
	})

	t.Run("after supernova, without flag activated at start, but in supernova genesis time", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		roundDuration := 10 * time.Millisecond

		supernovaRoundDuration := 5 * time.Millisecond
		supernovaStartRond := int64(5)
		supernovaGenesisTime := genesisTime.Add(time.Duration(supernovaStartRond) * roundDuration)

		args := createDefaultRoundArgs()
		args.RoundTimeDuration = roundDuration
		args.SupernovaTimeDuration = supernovaRoundDuration
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return flag == common.SupernovaRoundFlag && round >= uint64(supernovaStartRond)
			},
		}

		args.SupernovaStartRound = supernovaStartRond
		args.GenesisTimeStamp = genesisTime
		args.SupernovaGenesisTimeStamp = supernovaGenesisTime

		args.CurrentTimeStamp = supernovaGenesisTime

		rnd, _ := round.NewRound(args)

		assert.Equal(t, supernovaGenesisTime.UnixMilli(), rnd.TimeStamp().UnixMilli())
		index0 := rnd.Index()
		assert.Equal(t, supernovaRoundDuration, rnd.TimeDuration())
		assert.Equal(t, int64(5), index0)
	})

	t.Run("after supernova, without flag activated at start, but after supernova genesis time", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		roundDuration := 10 * time.Millisecond

		supernovaRoundDuration := 5 * time.Millisecond
		supernovaStartRond := int64(5)
		supernovaGenesisTime := genesisTime.Add(time.Duration(supernovaStartRond) * roundDuration)

		args := createDefaultRoundArgs()
		args.RoundTimeDuration = roundDuration
		args.SupernovaTimeDuration = supernovaRoundDuration
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return flag == common.SupernovaRoundFlag && round >= uint64(supernovaStartRond)
			},
		}

		args.SupernovaStartRound = supernovaStartRond
		args.GenesisTimeStamp = genesisTime
		args.SupernovaGenesisTimeStamp = supernovaGenesisTime

		currentTime := supernovaGenesisTime.Add(supernovaRoundDuration)
		args.CurrentTimeStamp = currentTime

		rnd, _ := round.NewRound(args)

		assert.Equal(t, supernovaRoundDuration, rnd.TimeDuration())
		assert.Equal(t, int64(6), rnd.Index())
		assert.Equal(t, currentTime.UnixMilli(), rnd.TimeStamp().UnixMilli())

		rnd.UpdateRound(genesisTime, supernovaGenesisTime.Add(2*supernovaRoundDuration))

		assert.Equal(t, supernovaRoundDuration, rnd.TimeDuration())
		assert.Equal(t, int64(7), rnd.Index())
		assert.Equal(t, currentTime.Add(supernovaRoundDuration).UnixMilli(), rnd.TimeStamp().UnixMilli())
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
	args.SupernovaGenesisTimeStamp = genesisTime.Add(10 * roundTimeDuration)
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

	t.Run("without supernova activated", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		syncTimerMock := &consensusMocks.SyncTimerMock{}

		startRound := int64(-1)
		args := createDefaultRoundArgs()
		args.GenesisTimeStamp = genesisTime
		args.SupernovaGenesisTimeStamp = genesisTime.Add(10 * roundTimeDuration)
		args.SyncTimer = syncTimerMock
		args.StartRound = startRound

		rnd, _ := round.NewRound(args)
		require.True(t, rnd.BeforeGenesis())

		time.Sleep(roundTimeDuration * 2)
		currentTime := time.Now()

		rnd.UpdateRound(genesisTime, currentTime)
		require.False(t, rnd.BeforeGenesis())
	})

	t.Run("with supernova activated", func(t *testing.T) {
		t.Parallel()

		roundTimeDuration := 10 * time.Millisecond
		supernovaRoundTimeDuration := 5 * time.Millisecond

		initTime := time.Now()
		genesisTime := initTime.Add(roundTimeDuration * 20)

		syncTimerMock := &consensusMocks.SyncTimerMock{}

		supernovaStartRound := int64(10)

		startRound := int64(0)
		args := createDefaultRoundArgs()
		args.GenesisTimeStamp = genesisTime
		args.SupernovaGenesisTimeStamp = genesisTime.Add(10 * roundTimeDuration)
		args.SyncTimer = syncTimerMock
		args.StartRound = startRound
		args.RoundTimeDuration = roundTimeDuration
		args.SupernovaTimeDuration = supernovaRoundTimeDuration
		args.SupernovaStartRound = supernovaStartRound
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return flag == common.SupernovaRoundFlag && round >= uint64(supernovaStartRound)
			},
		}

		rnd, _ := round.NewRound(args)
		require.True(t, rnd.BeforeGenesis())

		time.Sleep(roundTimeDuration * 10)
		rnd.UpdateRound(genesisTime, time.Now())
		require.True(t, rnd.BeforeGenesis())

		time.Sleep(roundTimeDuration * 11)
		rnd.UpdateRound(genesisTime, time.Now())
		require.False(t, rnd.BeforeGenesis())
	})
}

func TestRound_Concurrency(t *testing.T) {
	t.Parallel()

	t.Run("before supernova", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.GenesisTimeStamp = genesisTime

		rnd, err := round.NewRound(args)
		require.Nil(t, err)

		numOperations := 1000

		wg := sync.WaitGroup{}
		wg.Add(numOperations)

		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				switch idx % 6 {
				case 0:
					_ = rnd.BeforeGenesis()
				case 1:
					_ = rnd.Index()
				case 2:
					_ = rnd.RemainingTime(rnd.TimeStamp(), roundTimeDuration)
				case 3:
					rnd.RevertOneRound()
				case 4:
					_ = rnd.TimeDuration()
				case 5:
					rnd.UpdateRound(genesisTime, time.Now())
				default:
					assert.Fail(t, "should have not been called")
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})

	t.Run("after supernova", func(t *testing.T) {
		t.Parallel()

		genesisTime := time.Now()

		args := createDefaultRoundArgs()
		args.GenesisTimeStamp = genesisTime
		args.SupernovaGenesisTimeStamp = genesisTime
		args.SupernovaStartRound = 0

		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return flag == common.SupernovaRoundFlag && round > 0
			},
		}

		rnd, err := round.NewRound(args)
		require.Nil(t, err)

		numOperations := 1000

		wg := sync.WaitGroup{}
		wg.Add(numOperations)

		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				switch idx % 6 {
				case 0:
					_ = rnd.BeforeGenesis()
				case 1:
					_ = rnd.Index()
				case 2:
					_ = rnd.RemainingTime(rnd.TimeStamp(), roundTimeDuration)
				case 3:
					rnd.RevertOneRound()
				case 4:
					_ = rnd.TimeDuration()
				case 5:
					rnd.UpdateRound(genesisTime, time.Now())
				default:
					assert.Fail(t, "should have not been called")
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}

func TestRound_GetTimeStampForRound(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()
	supernovaGenesisTimeStamp := genesisTime.Add(10 * roundTimeDuration)

	roundTimeDuration := 10 * time.Millisecond
	supernovaRoundTimeDuration := 5 * time.Millisecond

	syncTimerMock := &consensusMocks.SyncTimerMock{}

	startRound := int64(0)
	supernovaStartRound := int64(10)

	args := createDefaultRoundArgs()
	args.GenesisTimeStamp = genesisTime
	args.SupernovaGenesisTimeStamp = supernovaGenesisTimeStamp
	args.SyncTimer = syncTimerMock
	args.StartRound = startRound
	args.RoundTimeDuration = roundTimeDuration
	args.SupernovaTimeDuration = supernovaRoundTimeDuration
	args.SupernovaStartRound = supernovaStartRound
	rnd, _ := round.NewRound(args)
	require.True(t, rnd.BeforeGenesis())

	roundTimeStamp := rnd.GetTimeStampForRound(0)
	expRoundTimeStamp := genesisTime.Add(0 * roundTimeDuration)
	require.Equal(t, uint64(expRoundTimeStamp.UnixMilli()), roundTimeStamp)

	roundTimeStamp = rnd.GetTimeStampForRound(10)
	expRoundTimeStamp = genesisTime.Add(10 * roundTimeDuration)
	require.Equal(t, uint64(expRoundTimeStamp.UnixMilli()), roundTimeStamp)

	roundTimeStamp = rnd.GetTimeStampForRound(20)
	expRoundTimeStamp = supernovaGenesisTimeStamp.Add((20 - 10) * supernovaRoundTimeDuration)
	require.Equal(t, uint64(expRoundTimeStamp.UnixMilli()), roundTimeStamp)

	roundTimeStamp = rnd.GetTimeStampForRound(1000)
	expRoundTimeStamp = supernovaGenesisTimeStamp.Add((1000 - 10) * supernovaRoundTimeDuration)
	require.Equal(t, uint64(expRoundTimeStamp.UnixMilli()), roundTimeStamp)
}
