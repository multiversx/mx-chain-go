package libp2p

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNilBlacklistHandler_ShouldWork(t *testing.T) {
	nbh := &nilBlacklistHandler{}

	assert.False(t, check.IfNil(nbh))
	assert.False(t, nbh.Has(""))
}
