package common

import (
	"time"

	"github.com/multiversx/mx-chain-go/errors"
)

const (
	genesisEpoch = 0

	// NumberOfSecondsInDay defines the number of seconds in a day
	NumberOfSecondsInDay = 86400
	// NumberOfMillisecondsInDay defines the number of milliseconds in a day
	NumberOfMillisecondsInDay = NumberOfSecondsInDay * 1000
)

const (
	minRoundDurationMS  = 200
	minRoundDurationSec = 1
)

// GetGenesisUnixTimestampFromStartTime returns genesis unix timestamp based on the
// provided time
func GetGenesisUnixTimestampFromStartTime(
	t time.Time,
	enableEpochsHandler EnableEpochsHandler,
) int64 {
	if enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, genesisEpoch) {
		return t.UnixMilli()
	}

	return t.Unix()
}

// GetGenesisStartTimeFromUnixTimestamp returns genesis time based on the provided
// unix timestamp
func GetGenesisStartTimeFromUnixTimestamp(
	unixTime int64,
	enableEpochsHandler EnableEpochsHandler,
) time.Time {
	if enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, genesisEpoch) {
		return time.UnixMilli(unixTime)
	}

	return time.Unix(unixTime, 0)
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

	return uint64(unitsInDay) / uint64(timeDurationToUnix(roundTime, enableEpochsHandler, epoch))
}

// timeDurationToUnix converts duration time to unix based on current configuration
func timeDurationToUnix(
	duration time.Duration,
	enableEpochsHandler EnableEpochsHandler,
	epoch uint32,
) int64 {
	if enableEpochsHandler.IsFlagEnabledInEpoch(SupernovaFlag, epoch) {
		return duration.Milliseconds()
	}

	return int64(duration.Seconds())
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

// RoundToNearestMinute rounds the given time to the nearest minute
func RoundToNearestMinute(
	t time.Time,
) time.Time {
	return t.Add(1 * time.Minute).Round(time.Minute)
}
