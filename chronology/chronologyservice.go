package chronology

import "time"

type IChronologyService interface {
	GetRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time) *Round
}
