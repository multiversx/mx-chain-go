package common

import (
	"time"

	"github.com/multiversx/mx-chain-go/errors"
)

const (
	// NumberOfSecondsInDay defines the number of seconds in a day
	NumberOfSecondsInDay = 86400
	// NumberOfMillisecondsInDay defines the number of milliseconds in a day
	NumberOfMillisecondsInDay = NumberOfSecondsInDay * 1000
)

const (
	minRoundDurationMS  = 500
	minRoundDurationSec = 1
)

// TimeToUnix returns the time to unix based on current configuration
func TimeToUnix(
	t time.Time,
	enableEpochsHandler EnableEpochsHandler,
) int64 {
	if enableEpochsHandler.IsFlagEnabled(SupernovaFlag) {
		return t.UnixMilli()
	}

	return t.Unix()
}

// TimeToUnixInEpoch returns the time to unix based on current configuration
func TimeToUnixInEpoch(
	t time.Time,
	enableEpochsHandler EnableEpochsHandler,
	epoch uint32,
) int64 {
	if enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, epoch) {
		return t.UnixMilli()
	}

	return t.Unix()
}

// UnixToTime converts int64 to time based on current configuration
func UnixToTime(
	unixTime int64,
	enableEpochsHandler EnableEpochsHandler,
	epoch uint32,
) time.Time {
	if enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, epoch) {
		return time.UnixMilli(unixTime)
	}

	return time.Unix(unixTime, 0)
}

// TimeDurationToUnix converts duration time to unix based on current configuration
func TimeDurationToUnix(
	duration time.Duration,
	enableEpochsHandler EnableEpochsHandler,
	epoch uint32,
) int64 {
	if enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, epoch) {
		return duration.Milliseconds()
	}

	return int64(duration.Seconds())
}

// CheckRoundDuration checks round duration based on current configuration
func CheckRoundDuration(
	roundDuration uint64,
	enableEpochsHandler EnableEpochsHandler,
) error {
	if enableEpochsHandler.IsFlagEnabled(SupernovaFlag) {
		return checkRoundDurationMilliSec(roundDuration)
	}

	return checkRoundDurationSec(roundDuration)
}

func checkRoundDurationSec(roundDuration uint64) error {
	roundDurationSec := roundDuration / 1000
	if roundDurationSec < minRoundDurationSec {
		return errors.ErrInvalidRoundDuration
	}

	return nil
}

func checkRoundDurationMilliSec(roundDuration uint64) error {
	if roundDuration < minRoundDurationMS {
		return errors.ErrInvalidRoundDuration
	}

	return nil
}

// ComputeRoundsPerDay computes the rounds per day based on current configuration
func ComputeRoundsPerDay(
	roundTime time.Duration,
	enableEpochsHandler EnableEpochsHandler,
	epoch uint32,
) uint64 {
	unitsInDay := getUnitsPerDay(enableEpochsHandler, epoch)

	return uint64(unitsInDay) / uint64(TimeDurationToUnix(roundTime, enableEpochsHandler, epoch))
}

func getUnitsPerDay(
	enableEpochsHandler EnableEpochsHandler,
	epoch uint32,
) int {
	if enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, epoch) {
		return NumberOfMillisecondsInDay
	}

	return NumberOfSecondsInDay
}
