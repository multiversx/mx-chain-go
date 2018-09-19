package chronology

import "time"

type IChronologyService interface {
	GetRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time) *Round
	GetRoundState(round *Round, timeStamp time.Time) RoundState
	GetRoundStateName(roundState RoundState) string
}
