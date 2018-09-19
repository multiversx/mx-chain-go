package consensus

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

func TestAnswerType(t *testing.T) {

	if AT_AGREE != 0 || AT_DISAGREE != 1 || AT_NOT_ANSWERED != 2 || AT_NOT_AVAILABLE != 3 {
		t.Fatal("Wrong values in answer type enum")
	}
}

func TestComputeLeader(t *testing.T) {

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	r := chronology.ChronologyServiceImpl{}.GetRoundFromDateTime(genesisRoundTimeStamp, time.Now(), time.Duration(4*time.Second))

	nodes := []string{"1", "2", "3"}

	node, err := ConsensusServiceImpl{}.ComputeLeader(nodes, r)

	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(node)
}

func TestNodeLeader(t *testing.T) {

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	r := chronology.ChronologyServiceImpl{}.GetRoundFromDateTime(genesisRoundTimeStamp, time.Now(), time.Duration(4*time.Second))

	node := "3"
	nodes := []string{"1", "2", "3"}

	bIsLeader, err := ConsensusServiceImpl{}.IsNodeLeader(node, nodes, r)

	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(bIsLeader)
}
