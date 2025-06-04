package common_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/require"
)

func TestTimeToUnix(t *testing.T) {
	t.Parallel()

	t.Run("with supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.SupernovaFlag && epoch == 0
			},
		}

		tt := time.Now()

		res := common.GetGenesisUnixTimestampFromStartTime(tt, enableEpochsHandler)
		require.Equal(t, tt.UnixMilli(), res)
	})

	t.Run("without supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag != common.SupernovaFlag
			},
		}

		tt := time.Now()

		res := common.GetGenesisUnixTimestampFromStartTime(tt, enableEpochsHandler)
		require.Equal(t, tt.Unix(), res)
	})
}

func TestTimeToUnixInEpoch(t *testing.T) {
	t.Parallel()

	t.Run("with supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.SupernovaFlag && epoch == 0
			},
		}

		tt := time.Now().Truncate(time.Millisecond)

		unixTimeStamp := common.GetGenesisUnixTimestampFromStartTime(tt, enableEpochsHandler)
		require.Equal(t, tt.UnixMilli(), unixTimeStamp)

		timeFromUnixTimeStamp := common.GetGenesisStartTimeFromUnixTimestamp(unixTimeStamp, enableEpochsHandler)
		require.Equal(t, tt, timeFromUnixTimeStamp)
	})

	t.Run("without supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag != common.SupernovaFlag && epoch == 0
			},
		}

		tt := time.Now().Truncate(time.Second)

		unixTimeStamp := common.GetGenesisUnixTimestampFromStartTime(tt, enableEpochsHandler)
		require.Equal(t, tt.Unix(), unixTimeStamp)

		timeFromUnixTimeStamp := common.GetGenesisStartTimeFromUnixTimestamp(unixTimeStamp, enableEpochsHandler)
		require.Equal(t, tt, timeFromUnixTimeStamp)
	})
}

func TestTimeDurationToUnix(t *testing.T) {
	t.Parallel()

	expEpoch := uint32(2)

	t.Run("with supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.SupernovaFlag && epoch == expEpoch
			},
		}

		dur := 200 * time.Millisecond

		unix := common.TimeDurationToUnix(dur, enableEpochsHandler, expEpoch)
		require.Equal(t, int64(200), unix)

		dur = 2 * time.Second

		unix = common.TimeDurationToUnix(dur, enableEpochsHandler, expEpoch)
		require.Equal(t, int64(2000), unix)
	})

	t.Run("without supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag != common.SupernovaFlag && epoch == expEpoch
			},
		}

		dur := 200 * time.Millisecond

		unix := common.TimeDurationToUnix(dur, enableEpochsHandler, expEpoch)
		require.Equal(t, int64(0), unix)

		dur = 2 * time.Second

		unix = common.TimeDurationToUnix(dur, enableEpochsHandler, expEpoch)
		require.Equal(t, int64(2), unix)
	})
}

func TestCheckRoundDuration(t *testing.T) {
	t.Parallel()

	t.Run("with supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SupernovaFlag
			},
		}

		err := common.CheckRoundDuration(uint64(common.MinRoundDurationMS-1), enableEpochsHandler)
		require.Equal(t, errors.ErrInvalidRoundDuration, err)

		err = common.CheckRoundDuration(uint64(0), enableEpochsHandler)
		require.Equal(t, errors.ErrInvalidRoundDuration, err)

		err = common.CheckRoundDuration(uint64(common.MinRoundDurationMS), enableEpochsHandler)
		require.Nil(t, err)
	})

	t.Run("without supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag != common.SupernovaFlag
			},
		}

		err := common.CheckRoundDuration(uint64(common.MinRoundDurationMS), enableEpochsHandler)
		require.Equal(t, errors.ErrInvalidRoundDuration, err)

		err = common.CheckRoundDuration(uint64(common.MinRoundDurationSec*1000-1), enableEpochsHandler)
		require.Equal(t, errors.ErrInvalidRoundDuration, err)

		err = common.CheckRoundDuration(uint64(common.MinRoundDurationSec*1000), enableEpochsHandler)
		require.Nil(t, err)
	})
}

func TestComputeRoundsPerDay(t *testing.T) {
	t.Parallel()

	expEpoch := uint32(2)

	t.Run("with supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.SupernovaFlag && epoch == expEpoch
			},
		}

		dur := 600 * time.Millisecond

		numRounds := common.ComputeRoundsPerDay(dur, enableEpochsHandler, expEpoch)
		expNumRounds := uint64(common.NumberOfMillisecondsInDay) / 600
		require.Equal(t, expNumRounds, numRounds)
	})

	t.Run("without supernova flag enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag != common.SupernovaFlag && epoch == expEpoch
			},
		}

		dur := 6 * time.Second

		numRounds := common.ComputeRoundsPerDay(dur, enableEpochsHandler, expEpoch)
		expNumRounds := uint64(common.NumberOfSecondsInDay) / 6
		require.Equal(t, expNumRounds, numRounds)
	})
}
