package networksharding

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNoSharderDistance(t *testing.T) {
	s := &NoSharder{}

	assert.False(t, check.IfNil(s))
	checkDistance(s, t)
}

func TestNoSharderOrdering(t *testing.T) {
	s := &NoSharder{}
	checkOrdering(s, t)
}
