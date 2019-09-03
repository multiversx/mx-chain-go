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

func TestCheckForNil_NilInterfaceShouldRetTrue(t *testing.T) {
	t.Parallel()

	assert.True(t, check.ForNil(nil))
}

func TestCheckForNil_NilUnderlyingLayerShouldRetTrue(t *testing.T) {
	t.Parallel()

	var tino *testInterfaceNilObj

	assert.True(t, check.ForNil(tino))
}

func TestCheckForNil_WithInstanceShouldRetFalse(t *testing.T) {
	t.Parallel()

	var tino *testInterfaceNilObj
	tino = &testInterfaceNilObj{}

	assert.False(t, check.ForNil(tino))
}
