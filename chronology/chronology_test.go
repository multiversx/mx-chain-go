package chronology

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/epoch"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

func TestRoundState(t *testing.T) {

	if round.RS_START_ROUND != 1 || round.RS_BLOCK != 2 || round.RS_END_ROUND != 7 {
		t.Fatal("Wrong values in round state enum")
	}
}

func TestEpochBehaviour(t *testing.T) {

	e := epoch.New(1, time.Now())
	spew.Dump(e)
}

func TestRoundBehaviour(t *testing.T) {

	rs := GetRounderService()

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := rs.CreateRoundTimeDivision(duration)

	r := rs.CreateRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division)

	roundState := rs.GetRoundStateFromDateTime(&r, time.Now())

	if roundState < round.RS_START_ROUND || roundState > round.RS_END_ROUND {
		t.Fatal("Wrong round state")
	}

	spew.Dump(r)
	spew.Dump(roundState)
	spew.Dump(rs.GetRoundStateName(roundState))
}

func TestComputeLeader(t *testing.T) {

	var chr Chronology
	rs := GetRounderService()

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := rs.CreateRoundTimeDivision(duration)

	round := rs.CreateRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division)

	nodes := []string{"1", "2", "3"}

	node, err := chr.ComputeLeader(nodes, &round)

	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(node)
}

func TestNodeLeader(t *testing.T) {

	var chr Chronology
	rs := GetRounderService()

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := rs.CreateRoundTimeDivision(duration)

	round := rs.CreateRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division)

	node := "3"
	nodes := []string{"1", "2", "3"}

	bIsLeader, err := chr.IsNodeLeader(node, nodes, &round)

	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(bIsLeader)
}
