package disabled

import "time"

// RoundHandler implements the RoundHandler interface but does nothing as it is disabled
type RoundHandler struct {
}

// BeforeGenesis -
func (rh *RoundHandler) BeforeGenesis() bool {
	return false
}

// Index -
func (rh *RoundHandler) Index() int64 {
	return 0
}

// TimeDuration -
func (rh *RoundHandler) TimeDuration() time.Duration {
	return 0
}

// TimeStamp -
func (rh *RoundHandler) TimeStamp() time.Time {
	return time.Unix(0, 0)
}

// UpdateRound -
func (rh *RoundHandler) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
}

// RemainingTime -
func (rh *RoundHandler) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	return 0
}

// IncrementIndex -
func (rh *RoundHandler) IncrementIndex() {
}

// GetTimeStampForRound -
func (rh *RoundHandler) GetTimeStampForRound(round uint64) uint64 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (rh *RoundHandler) IsInterfaceNil() bool {
	return rh == nil
}
