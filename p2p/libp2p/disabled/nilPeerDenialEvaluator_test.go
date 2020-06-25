package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNilPeerDenialEvaluator_ShouldWork(t *testing.T) {
	nbh := &NilPeerDenialEvaluator{}

	assert.False(t, check.IfNil(nbh))
	assert.False(t, nbh.IsDenied(""))
}
