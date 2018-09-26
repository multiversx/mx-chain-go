package chronology

import (
	epoch "github.com/ElrondNetwork/elrond-go-sandbox/chronology/epoch"
	round "github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"time"
)

type Epocher interface {
	Print(epoch *epoch.Epoch)
}

type Rounder interface {
	CreateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration, roundTimeDivision []time.Duration) *round.Round
	UpdateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, round *round.Round)
	CreateRoundTimeDivision(time.Duration) []time.Duration
	GetRoundStateFromDateTime(round *round.Round, timeStamp time.Time) round.RoundState
	GetRoundStateName(roundState round.RoundState) string
}
