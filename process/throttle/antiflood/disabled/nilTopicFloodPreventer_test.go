package disabled

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewNilTopicFloodPreventer(t *testing.T) {
	t.Parallel()

	ntfp := NewNilTopicFloodPreventer()

	assert.False(t, check.IfNil(ntfp))
}

func TestNilTopicFloodPreventer_FunctionsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have panicked %v", r))
		}
	}()

	ntfp := NewNilTopicFloodPreventer()

	ntfp.ResetForTopic("")
	ntfp.SetMaxMessagesForTopic("", 0)
	assert.Nil(t, ntfp.IncreaseLoad("", "", math.MaxUint32))
}
