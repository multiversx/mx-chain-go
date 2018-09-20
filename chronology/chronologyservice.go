package chronology

import "time"

type IChronologyService interface {
	CreateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration, roundTimeDivision []time.Duration) *Round
	UpdateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, round *Round)
	CreateRoundTimeDivision(time.Duration) []time.Duration
	GetRoundStateFromDateTime(round *Round, timeStamp time.Time) RoundState
	GetRoundStateName(roundState RoundState) string
}
