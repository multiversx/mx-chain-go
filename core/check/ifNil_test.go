package check_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

type testInterfaceNilObj struct {
}

func (tino *testInterfaceNilObj) IsInterfaceNil() bool {
	return tino == nil
}

func TestCheckIfNil_NilInterfaceShouldRetTrue(t *testing.T) {
	t.Parallel()

	assert.True(t, check.IfNil(nil))
	assert.True(t, check.IfNilReflect(nil))
}

func TestCheckIfNil_NilUnderlyingLayerShouldRetTrue(t *testing.T) {
	t.Parallel()

	var tino *testInterfaceNilObj

	assert.True(t, check.IfNil(tino))
	assert.True(t, check.IfNilReflect(tino))
}

func TestCheckIfNil_WithInstanceShouldRetFalse(t *testing.T) {
	t.Parallel()

	tino := &testInterfaceNilObj{}

	assert.False(t, check.IfNil(tino))
	assert.False(t, check.IfNilReflect(tino))
}

func BenchmarkIfNilChecker(b *testing.B) {
	var nr *testInterfaceNilObj
	r := testInterfaceNilObj{}
	list := []check.NilInterfaceChecker{nil, &r, nr}

	for n := 0; n < b.N; n++ {
		for _, i := range list {
			check.IfNil(i)
		}
	}
}

func BenchmarkCheckIsItfReflect(b *testing.B) {
	var ns *struct{}
	s := struct{}{}

	list := []interface{}{nil, ns, &s}
	for n := 0; n < b.N; n++ {
		for _, i := range list {
			check.IfNilReflect(i)
		}
	}
}
