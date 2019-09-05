package check_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

type testInterfaceNilObj struct {
}

func (tino *testInterfaceNilObj) IsInterfaceNil() bool {
	if tino == nil {
		return true
	}
	return false
}

func TestCheckIfNil_NilInterfaceShouldRetTrue(t *testing.T) {
	t.Parallel()

	assert.True(t, check.IfNil(nil))
}

func TestCheckIfNil_NilUnderlyingLayerShouldRetTrue(t *testing.T) {
	t.Parallel()

	var tino *testInterfaceNilObj

	assert.True(t, check.IfNil(tino))
}

func TestCheckIfNil_WithInstanceShouldRetFalse(t *testing.T) {
	t.Parallel()

	var tino *testInterfaceNilObj
	tino = &testInterfaceNilObj{}

	assert.False(t, check.IfNil(tino))
}
