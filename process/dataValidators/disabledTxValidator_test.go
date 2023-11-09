package dataValidators

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledTxValidator(t *testing.T) {
	t.Parallel()

	dtv := NewDisabledTxValidator()

	assert.False(t, check.IfNil(dtv))
}

func TestDisabledTxValidator_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r, fmt.Sprintf("should have not paniced: %v", r))
	}()

	dtv := NewDisabledTxValidator()
	assert.Nil(t, dtv.CheckTxValidity(nil))
	assert.Nil(t, dtv.CheckTxWhiteList(nil))
}
