package disabled

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestShardIDProvider_MethodsDoNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	s := NewShardIDProvider()
	assert.Zero(t, s.ComputeId([]byte{}))
	assert.Zero(t, s.NumberOfShards())
	assert.Nil(t, s.GetShardIDs())
	assert.False(t, check.IfNil(s))
}
