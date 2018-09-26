package chronology

import (
	epoch "github.com/ElrondNetwork/elrond-go-sandbox/chronology/epoch"
	round "github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

func TestRoundState(t *testing.T) {

	if round.RS_START_ROUND != 1 || round.RS_PROPOSE_BLOCK != 2 || round.RS_END_ROUND != 7 {
		t.Fatal("Wrong values in round state enum")
	}
}

func TestEpochBehaviour(t *testing.T) {

	es := GetEpocherService()

	e := epoch.New(1, time.Now())
	es.Print(&e)
}

func TestRoundBehaviour(t *testing.T) {

	rs := GetRounderService()

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := rs.CreateRoundTimeDivision(duration)

	r := rs.CreateRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division)

	roundState := rs.GetRoundStateFromDateTime(r, time.Now())

	if roundState < round.RS_START_ROUND || roundState > round.RS_END_ROUND {
		t.Fatal("Wrong round state")
	}

	spew.Dump(r)
	spew.Dump(roundState)
	spew.Dump(rs.GetRoundStateName(roundState))
}
