package core_test

import (
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestMaxInt32ShouldReturnA(t *testing.T) {
	a := int32(-1)
	b := int32(-2)
	assert.Equal(t, a, core.MaxInt32(a, b))
}

func TestMaxInt32ShouldReturnB(t *testing.T) {
	a := int32(-2)
	b := int32(-1)
	assert.Equal(t, b, core.MaxInt32(a, b))
}

func TestMinInt32ShouldReturnB(t *testing.T) {
	a := int32(-1)
	b := int32(-2)
	assert.Equal(t, b, core.MinInt32(a, b))
}

func TestMinInt32ShouldReturnA(t *testing.T) {
	a := int32(-2)
	b := int32(-1)
	assert.Equal(t, a, core.MinInt32(a, b))
}

func TestMaxUint32ShouldReturnA(t *testing.T) {
	a := uint32(11)
	b := uint32(10)
	assert.Equal(t, a, core.MaxUint32(a, b))
}

func TestMaxUint32ShouldReturnB(t *testing.T) {
	a := uint32(10)
	b := uint32(11)
	assert.Equal(t, b, core.MaxUint32(a, b))
}

func TestMinUint32ShouldReturnB(t *testing.T) {
	a := uint32(11)
	b := uint32(10)
	assert.Equal(t, b, core.MinUint32(a, b))
}

func TestMinUint32ShouldReturnA(t *testing.T) {
	a := uint32(10)
	b := uint32(11)
	assert.Equal(t, a, core.MinUint32(a, b))
}

func TestMaxUint64ShouldReturnA(t *testing.T) {
	a := uint64(11)
	b := uint64(10)
	assert.Equal(t, a, core.MaxUint64(a, b))
}

func TestMaxUint64ShouldReturnB(t *testing.T) {
	a := uint64(10)
	b := uint64(11)
	assert.Equal(t, b, core.MaxUint64(a, b))
}

func TestMinUint64ShouldReturnB(t *testing.T) {
	a := uint64(11)
	b := uint64(10)
	assert.Equal(t, b, core.MinUint64(a, b))
}

func TestMinUint64ShouldReturnA(t *testing.T) {
	a := uint64(10)
	b := uint64(11)
	assert.Equal(t, a, core.MinUint64(a, b))
}

func TestMaxInt64ShouldReturnA(t *testing.T) {
	a := int64(11)
	b := int64(10)
	assert.Equal(t, a, core.MaxInt64(a, b))
}

func TestMaxInt64ShouldReturnB(t *testing.T) {
	a := int64(10)
	b := int64(11)
	assert.Equal(t, b, core.MaxInt64(a, b))
}

func TestMinInt64ShouldReturnB(t *testing.T) {
	a := int64(11)
	b := int64(10)
	assert.Equal(t, b, core.MinInt64(a, b))
}

func TestMinInt64ShouldReturnA(t *testing.T) {
	a := int64(10)
	b := int64(11)
	assert.Equal(t, a, core.MinInt64(a, b))
}

func TestMinIntShouldReturnB(t *testing.T) {
	a := 11
	b := 10
	assert.Equal(t, b, core.MinInt(a, b))
}

func TestMinIntShouldReturnA(t *testing.T) {
	a := 10
	b := 11
	assert.Equal(t, a, core.MinInt(a, b))
}

func TestMaxIntShouldReturnA(t *testing.T) {
	a := 11
	b := 10
	assert.Equal(t, a, core.MaxInt(a, b))
}

func TestMaxIntShouldReturnB(t *testing.T) {
	a := 10
	b := 11
	assert.Equal(t, b, core.MaxInt(a, b))
}

func TestMaxFloat64(t *testing.T) {
	tests := []struct {
		a      float64
		b      float64
		result float64
	}{
		{
			a:      5.1,
			b:      5.0,
			result: 5.1,
		},
		{
			a:      5.0,
			b:      5.1,
			result: 5.1,
		},
		{
			a:      -5.1,
			b:      -5.0,
			result: -5.0,
		},
		{
			a:      37.0,
			b:      36.999999,
			result: 37.0,
		},
		{
			a:      -5.3,
			b:      0,
			result: 0,
		},
		{
			a:      5.66,
			b:      5.66,
			result: 5.66,
		},
	}

	for _, tt := range tests {
		res := core.MaxFloat64(tt.a, tt.b)
		assert.Equal(t, tt.result, res)
	}
}

func TestSafeAddUint64(t *testing.T) {
	a := uint64(10)
	b := uint64(10)

	c, err := core.SafeAddUint64(a, b)
	assert.Nil(t, err)
	assert.Equal(t, a+b, c)

	a = math.MaxUint64
	c, err = core.SafeAddUint64(a, b)
	assert.Equal(t, err, core.ErrAdditionOverflow)
	assert.Equal(t, uint64(0), c)
}

func TestSafeSubUint64(t *testing.T) {
	a := uint64(20)
	b := uint64(10)

	c, err := core.SafeSubUint64(a, b)
	assert.Nil(t, err)
	assert.Equal(t, a-b, c)

	c, err = core.SafeSubUint64(b, a)
	assert.Equal(t, err, core.ErrSubtractionOverflow)
	assert.Equal(t, uint64(0), c)
}
