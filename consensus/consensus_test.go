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
