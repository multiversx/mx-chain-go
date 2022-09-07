package disabled

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledValidatorInfoResolver(t *testing.T) {
	t.Parallel()

	resolver := NewDisabledValidatorInfoResolver()
	assert.False(t, check.IfNil(resolver))
}

func Test_validatorInfoResolver_SetResolverDebugHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
		}
	}()

	resolver := NewDisabledValidatorInfoResolver()

	err := resolver.RequestDataFromHash(nil, 0)
	assert.Nil(t, err)

	err = resolver.RequestDataFromHashArray(nil, 0)
	assert.Nil(t, err)

	err = resolver.SetResolverDebugHandler(nil)
	assert.Nil(t, err)

	value1, value2 := resolver.NumPeersToQuery()
	assert.Zero(t, value1)
	assert.Zero(t, value2)

	err = resolver.Close()
	assert.Nil(t, err)

	resolver.SetNumPeersToQuery(100, 100)
}
