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

	dbh := &BlacklistHandler{}
	assert.False(t, check.IfNil(dbh))

	val := dbh.Has("a")
	assert.False(t, val)

	err := dbh.Add("")
	assert.Nil(t, err)

	err = dbh.AddWithSpan("", 0)
	assert.Nil(t, err)

	dbh.Sweep()
}
