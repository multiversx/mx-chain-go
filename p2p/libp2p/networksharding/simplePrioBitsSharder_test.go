package networksharding

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestSimplePrioSharderDistance(t *testing.T) {
	s := &SimplePrioBitsSharder{}

	assert.False(t, check.IfNil(s))
	checkDistance(s, t)
}

func TestSimplePrioBitsSharderOrdering(t *testing.T) {
	s := &SimplePrioBitsSharder{}
	checkOrdering(s, t)
}
