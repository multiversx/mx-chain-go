package chronology

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

func TestRoundState(t *testing.T) {

	if RS_START_ROUND != 1 || RS_PROPOSE_BLOCK != 2 || RS_END_ROUND != 7 {
		t.Fatal("Wrong values in round state enum")
	}
}

func TestEpochBehaviour(t *testing.T) {
	e := NewEpoch(1, time.Now())
	EpochServiceImpl{}.Print(&e)
}

func TestRoundBehaviour(t *testing.T) {

	var csi ChronologyServiceImpl

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := csi.CreateRoundTimeDivision(duration)

	round := csi.CreateRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division)

	roundState := csi.GetRoundStateFromDateTime(round, time.Now())

	if roundState < RS_START_ROUND || roundState > RS_END_ROUND {
		t.Fatal("Wrong round state")
	}

	spew.Dump(round)
	spew.Dump(roundState)
	spew.Dump(csi.GetRoundStateName(roundState))
}
