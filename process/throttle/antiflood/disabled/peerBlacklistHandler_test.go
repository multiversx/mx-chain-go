package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestBlacklistHandler_ShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r, "this shouldn't panic")
	}()

	pdbh := &PeerBlacklistHandler{}
	assert.False(t, check.IfNil(pdbh))

	val := pdbh.Has("a")
	assert.False(t, val)

	err := pdbh.Add("")
	assert.Nil(t, err)

	err = pdbh.AddWithSpan("", 0)
	assert.Nil(t, err)

	pdbh.Sweep()
}
