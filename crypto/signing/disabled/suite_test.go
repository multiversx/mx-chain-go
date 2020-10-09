package disabled

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledSuite_MethodsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	ds := NewDisabledSuite()

	recovPoint, err := ds.CreatePointForScalar(nil)
	assert.Equal(t, &disabledPoint{}, recovPoint)
	assert.Nil(t, err)

	recovScalar, recovPoint := ds.CreateKeyPair()
	assert.Equal(t, &disabledScalar{}, recovScalar)
	assert.Equal(t, &disabledPoint{}, recovPoint)

	assert.Equal(t, Disabled, ds.String())
	assert.Equal(t, scalarLen, ds.ScalarLen())
	assert.Equal(t, &disabledScalar{}, ds.CreateScalar())
	assert.Equal(t, pointLen, ds.PointLen())
	assert.Equal(t, &disabledPoint{}, ds.CreatePoint())
	assert.False(t, check.IfNil(ds))
	assert.Nil(t, ds.RandomStream())
	assert.Nil(t, ds.CheckPointValid(nil))
	assert.Nil(t, ds.GetUnderlyingSuite())
}
