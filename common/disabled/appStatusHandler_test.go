package disabled

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestAppStatusHandler_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	ash := NewAppStatusHandler()
	assert.False(t, check.IfNil(ash))

	ash.AddUint64("key", uint64(0))
	ash.Increment("key")
	ash.Decrement("key")
	ash.SetInt64Value("key", int64(1))
	ash.SetUInt64Value("key", uint64(2))
	ash.SetStringValue("key", "true")
	ash.Close()
}
