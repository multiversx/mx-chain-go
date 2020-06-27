package check_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

type testStructIsZeroFloat64 struct {
	value          float64
	epsilon        float64
	expectedResult bool
}

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

func TestIsZeroFloat64(t *testing.T) {
	t.Parallel()

	testValues := []testStructIsZeroFloat64{
		{
			value:          0,
			epsilon:        0,
			expectedResult: true,
		},
		{
			value:          0,
			epsilon:        1,
			expectedResult: true,
		},
		{
			value:          0.1,
			epsilon:        1,
			expectedResult: true,
		},
		{
			value:          -0.1,
			epsilon:        1,
			expectedResult: true,
		},
		{
			value:          1,
			epsilon:        1,
			expectedResult: true,
		},
		{
			value:          1,
			epsilon:        0.01,
			expectedResult: false,
		},
		{
			value:          0,
			epsilon:        -1,
			expectedResult: true,
		},
		{
			value:          0.1,
			epsilon:        -1,
			expectedResult: true,
		},
		{
			value:          -0.1,
			epsilon:        -1,
			expectedResult: true,
		},
		{
			value:          -1,
			epsilon:        -1,
			expectedResult: true,
		},
		{
			value:          -1,
			epsilon:        0.01,
			expectedResult: false,
		},
		{
			value:          0.000000000000000000009,
			epsilon:        0.00000000000000000001,
			expectedResult: true,
		},
	}

	for _, test := range testValues {
		assert.Equal(t, test.expectedResult, check.IsZeroFloat64(test.value, test.epsilon),
			fmt.Sprintf("test failed for value: %.4f, epsilon %.4f and expected value %v", test.value, test.epsilon, test.expectedResult))
	}
}
