package singlesig

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabled_MethodsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	d := &Disabled{}

	recovBytes, err := d.Sign(nil, nil)
	assert.Equal(t, []byte(signature), recovBytes)
	assert.Nil(t, err)

	assert.Nil(t, d.Verify(nil, nil, nil))
	assert.False(t, check.IfNil(d))
}
