package disabled

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestBlacklistHandler_ShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r, "this shouldn't panic")
	}()

	pbc := &PeerBlacklistCacher{}
	assert.False(t, check.IfNil(pbc))

	val := pbc.Has("a")
	assert.False(t, val)

	err := pbc.Upsert("", time.Second)
	assert.Nil(t, err)

	pbc.Sweep()
}
