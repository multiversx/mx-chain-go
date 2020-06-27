package disabled

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNilPeerDenialEvaluator_ShouldWork(t *testing.T) {
	nbh := &NilPeerDenialEvaluator{}

	assert.False(t, check.IfNil(nbh))
	assert.Nil(t, nbh.UpsertPeerID("", time.Second))
	assert.False(t, nbh.IsDenied(""))
}
