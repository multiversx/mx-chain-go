package networksharding

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledSharder(t *testing.T) {
	ds := NewDisabledSharder()

	assert.False(t, check.IfNil(ds))
	assert.Equal(t, 0, len(ds.ComputeEvictionList(nil)))
	assert.False(t, ds.Has("", nil))
	assert.Nil(t, ds.SetPeerShardResolver(nil))
}
