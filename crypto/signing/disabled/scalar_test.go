package disabled

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledScalar_MethodsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	ds := &disabledScalar{}

	recovBytes, err := ds.MarshalBinary()
	assert.Equal(t, []byte(marshaledScalar), recovBytes)
	assert.Nil(t, err)

	recovBool, err := ds.Equal(nil)
	assert.False(t, recovBool)
	assert.Nil(t, err)

	recovScalar, err := ds.Add(nil)
	assert.Equal(t, ds, recovScalar)
	assert.Nil(t, err)

	recovScalar, err = ds.Sub(nil)
	assert.Equal(t, ds, recovScalar)
	assert.Nil(t, err)

	recovScalar, err = ds.Mul(nil)
	assert.Equal(t, ds, recovScalar)
	assert.Nil(t, err)

	recovScalar, err = ds.Div(nil)
	assert.Equal(t, ds, recovScalar)
	assert.Nil(t, err)

	recovScalar, err = ds.Inv(nil)
	assert.Equal(t, ds, recovScalar)
	assert.Nil(t, err)

	recovScalar, err = ds.Pick()
	assert.Equal(t, ds, recovScalar)
	assert.Nil(t, err)

	recovScalar, err = ds.SetBytes(nil)
	assert.Equal(t, ds, recovScalar)
	assert.Nil(t, err)

	assert.Nil(t, ds.UnmarshalBinary(nil))
	assert.Nil(t, ds.Set(nil))
	assert.Equal(t, ds, ds.Clone())
	ds.SetInt64(0)
	assert.Equal(t, ds, ds.Zero())
	assert.Equal(t, ds, ds.Neg())
	assert.Equal(t, ds, ds.One())
	assert.Nil(t, ds.GetUnderlyingObj())
	assert.False(t, check.IfNil(ds))
}
