package resolver

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledInterceptorResolver(t *testing.T) {
	t.Parallel()

	dir := NewDisabledInterceptorResolver()
	assert.False(t, check.IfNil(dir))

	assert.False(t, dir.Enabled())
	dir.LogReceivedHash("", nil)
	dir.LogProcessedHash("", nil, nil)
	dir.LogRequestedData("", nil, 0, 0)
	dir.LogFailedToResolveData("", nil, nil)
	assert.Equal(t, 0, len(dir.Query("*")))
}
