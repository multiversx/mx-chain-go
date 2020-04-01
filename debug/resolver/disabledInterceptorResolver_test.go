package resolver

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledInterceptorResolver(t *testing.T) {
	t.Parallel()

	dir, err := NewDisabledInterceptorResolver()
	assert.False(t, check.IfNil(dir))
	assert.Nil(t, err)

	assert.False(t, dir.Enabled())
	dir.ReceivedHash("", nil)
	dir.ProcessedHash("", nil, nil)
	dir.RequestedData("", nil, 0, 0)
	assert.Equal(t, 0, len(dir.Query("*")))
}
