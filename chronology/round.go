package chronology

import "time"

type Round struct {
	index             int64
	startTimeStamp    time.Time
	roundTimeDuration time.Duration
}

func NewRound(index int64, startTimeStamp time.Time, roundTimeDuration time.Duration) Round {

	r := Round{index, startTimeStamp, roundTimeDuration}
	return r
}

func (r *Round) SetIndex(index int64) {
	r.index = index
}

func (r *Round) GetIndex() int64 {
	return r.index
}

func (r *Round) SetStartTimeStamp(startTimeStamp time.Time) {
	r.startTimeStamp = startTimeStamp
}

func (r *Round) GetStartTimeStamp() time.Time {
	return r.startTimeStamp
}

func (r *Round) SetRoundTimeDuration(roundTimeDuration time.Duration) {
	r.roundTimeDuration = roundTimeDuration
}

func (r *Round) GetRoundTimeDuration() time.Duration {
	return r.roundTimeDuration
}
