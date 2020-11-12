package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledFloodPreventer_ShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r, "this shouldn't panic")
	}()

	daf := &disabledFloodPreventer{}
	assert.False(t, check.IfNil(daf))

	_ = daf.IncreaseLoad("test", 10)
	_ = daf.IncreaseLoadGlobal("test", 10)
	daf.Reset()
}
