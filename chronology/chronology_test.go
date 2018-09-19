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

func TestRoundBehaviour(t *testing.T) {

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	r := ChronologyServiceImpl{}.GetRoundFromDateTime(genesisRoundTimeStamp, time.Now(), time.Duration(4*time.Second))

	roundState := ChronologyServiceImpl{}.GetRoundState(r, time.Now())

	if roundState < RS_START_ROUND || roundState > RS_END_ROUND {
		t.Fatal("Wrong round state")
	}

	spew.Dump(r)
	spew.Dump(roundState)
	spew.Dump(ChronologyServiceImpl{}.GetRoundStateName(roundState))
}
