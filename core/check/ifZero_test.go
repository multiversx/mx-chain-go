package check_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	a uint64
	b uint32
	c uint
	d string
}

func TestCheckForZeroUintFields(t *testing.T) {
	zero := testStruct{}
	err := check.ForZeroUintFields(zero)
	assert.NotNil(t, err)

	zero = testStruct{
		a: 0,
		b: 1,
		c: 1,
		d: "",
	}
	err = check.ForZeroUintFields(zero)
	assert.NotNil(t, err)

	zero = testStruct{
		a: 1,
		b: 0,
		c: 1,
		d: "",
	}
	err = check.ForZeroUintFields(zero)
	assert.NotNil(t, err)

	zero = testStruct{
		a: 1,
		b: 1,
		c: 0,
		d: "",
	}
	err = check.ForZeroUintFields(zero)
	assert.NotNil(t, err)

	zero = testStruct{
		a: 1,
		b: 1,
		c: 1,
		d: "",
	}
	err = check.ForZeroUintFields(zero)
	assert.Nil(t, err)
}
