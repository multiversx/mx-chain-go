package chronology

import "testing"

func TestRoundState(t *testing.T) {

	if RS_START_ROUND != 0 || RS_PROPOSE_BLOCK != 1 || RS_END_ROUND != 2 {
		t.Fatal("Wrong values in round state enum")
	}
}
