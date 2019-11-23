package networksharding

import (
	"testing"
)

func TestNoSharderDistance(t *testing.T) {
	s := &NoSharder{}
	checkDistance(s, t)
}

func TestNoSharderOrdering(t *testing.T) {
	s := &NoSharder{}
	checkOrdering(s, t)
}
