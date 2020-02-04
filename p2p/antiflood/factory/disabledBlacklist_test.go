package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledBlacklistHandler_ShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r, "this shouldn't panic")
	}()

	dbh := &disabledBlacklistHandler{}
	assert.False(t, check.IfNil(dbh))

	_ = dbh.Has("a")
}
