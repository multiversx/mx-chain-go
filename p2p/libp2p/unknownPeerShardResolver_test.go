package libp2p

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestUnknownPeerShardResolver_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var upsr *unknownPeerShardResolver
	assert.True(t, check.IfNil(upsr))

	upsr = &unknownPeerShardResolver{}
	assert.False(t, check.IfNil(upsr))
}

func TestUnknownPeerShardResolver_ByIDShouldReturnUnknownId(t *testing.T) {
	t.Parallel()

	upsr := &unknownPeerShardResolver{}

	assert.Equal(t, sharding.UnknownShardId, upsr.ByID(""))
}
