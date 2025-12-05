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

		tt := time.Now().Truncate(time.Second)

		unixTimeStamp := common.GetGenesisUnixTimestampFromStartTime(tt, enableEpochsHandler)
		require.Equal(t, tt.UnixMilli(), unixTimeStamp)

		// should not work properly to get start time from unix timestamp as milliseconds
		timeFromUnixTimeStamp := common.GetGenesisStartTimeFromUnixTimestamp(unixTimeStamp, enableEpochsHandler)
		require.NotEqual(t, tt, timeFromUnixTimeStamp)
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

func TestRoundToNearestMinute(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Round up at 30 seconds",
			input:    time.Date(2025, 6, 13, 16, 8, 30, 0, time.UTC),
			expected: time.Date(2025, 6, 13, 16, 10, 0, 0, time.UTC), // 16:08:30 + 1m = 16:09:30 -> rounds to 16:10:00
		},
		{
			name:     "Round down at 29 seconds",
			input:    time.Date(2025, 6, 13, 16, 8, 29, 999999999, time.UTC),
			expected: time.Date(2025, 6, 13, 16, 9, 0, 0, time.UTC), // 16:08:29.999 + 1m = 16:09:29.999 -> rounds to 16:09:00
		},
		{
			name:     "Exact minute",
			input:    time.Date(2025, 6, 13, 16, 8, 0, 0, time.UTC),
			expected: time.Date(2025, 6, 13, 16, 9, 0, 0, time.UTC), // 16:08:00 + 1m = 16:09:00 -> rounds to 16:09:00
		},
		{
			name:     "Round up at 45 seconds",
			input:    time.Date(2025, 6, 13, 16, 8, 45, 500000000, time.UTC),
			expected: time.Date(2025, 6, 13, 16, 10, 0, 0, time.UTC), // 16:08:45.5 + 1m = 16:09:45.5 -> rounds to 16:10:00
		},
		{
			name:     "Different timezone",
			input:    time.Date(2025, 6, 13, 16, 8, 30, 0, time.FixedZone("EST", -5*60*60)),
			expected: time.Date(2025, 6, 13, 16, 10, 0, 0, time.FixedZone("EST", -5*60*60)), // 16:08:30 + 1m = 16:09:30 -> rounds to 16:10:00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.RoundToNearestMinute(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("RoundToNearestMinute(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
