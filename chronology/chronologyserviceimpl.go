package chronology

import (
	"fmt"
	"time"
)

var RoundTimeDuration time.Duration

type ChronologyServiceImpl struct {
}

func init() {

	RoundTimeDuration = time.Duration(4 * time.Second)
}

func (ChronologyServiceImpl) GetRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time) *Round {

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	if delta < 0 {
		fmt.Print("genesisRoundTimeStamp should be lower or equal to timestamp!\n")
		return nil
	}

	var r Round

	r.SetIndex(delta / RoundTimeDuration.Nanoseconds())
	r.SetStartTimeStamp(genesisRoundTimeStamp.Add(time.Duration(r.GetIndex() * RoundTimeDuration.Nanoseconds())))
	r.SetRoundTimeDuration(RoundTimeDuration)

	return &r
}
