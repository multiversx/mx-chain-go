package networksharding

import (
	"testing"
)

func TestNoSharderDistance(t *testing.T) {
	s := &noSharder{}
	checkDistance(s, t)
}

func TestNoSharderOrdering(t *testing.T) {
	s := &noSharder{}
	checkOrdering(s, t)
}
