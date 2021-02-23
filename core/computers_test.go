package core_test

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIsInRangeInclusive(t *testing.T) {
	// positive edges
	min := big.NewInt(0)
	max := big.NewInt(100)

	// within boundaries
	val := big.NewInt(5)
	inRange := core.IsInRangeInclusive(val, min, max)
	require.True(t, inRange)

	// outside boundaries low
	val = big.NewInt(-1)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.False(t, inRange)

	// outside boundaries high
	val = big.NewInt(110)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.False(t, inRange)

	// lower edge
	val = big.NewInt(0)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.True(t, inRange)

	// higher edge
	val = big.NewInt(100)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.True(t, inRange)

	// negative edges
	min = big.NewInt(-100)
	max = big.NewInt(-10)

	// within boundaries
	val = big.NewInt(-50)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.True(t, inRange)

	// outside boundaries low
	val = big.NewInt(-110)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.False(t, inRange)

	// outside boundaries high
	val = big.NewInt(-5)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.False(t, inRange)

	// lower edge
	val = big.NewInt(-100)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.True(t, inRange)

	// higher edge
	val = big.NewInt(-10)
	inRange = core.IsInRangeInclusive(val, min, max)
	require.True(t, inRange)
}

func TestIsInRangeInclusiveFloat64(t *testing.T) {
	// positive edges
	min := 0.001
	max := 100.001

	// within boundaries
	val := 5.001
	inRange := core.IsInRangeInclusiveFloat64(val, min, max)
	require.True(t, inRange)

	// outside boundaries low
	val = -1.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.False(t, inRange)

	// outside boundaries high
	val = 110.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.False(t, inRange)

	// lower edge
	val = 0.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.True(t, inRange)

	// higher edge
	val = 100.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.True(t, inRange)

	// negative edges
	min = -100.001
	max = -10.001

	// within boundaries
	val = -50.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.True(t, inRange)

	// outside boundaries low
	val = -110.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.False(t, inRange)

	// outside boundaries high
	val = -5.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.False(t, inRange)

	// lower edge
	val = -100.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.True(t, inRange)

	// higher edge
	val = -10.001
	inRange = core.IsInRangeInclusiveFloat64(val, min, max)
	require.True(t, inRange)
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

func TestGetPercentageNoLoss(t *testing.T) {
	a := "29815853976407917651"
	percentage := 0.1
	expected := "2981585397640791765"

	bigA, _ := big.NewInt(0).SetString(a, 10)
	bigExpected, _ := big.NewInt(0).SetString(expected, 10)
	result := core.GetIntTrimmedPercentageOfValue(bigA, percentage)
	require.Equal(t, bigExpected, result)

	a = "29815853976407917651"
	percentage = 0.01
	expected = "298158539764079176"
	bigA, _ = big.NewInt(0).SetString(a, 10)
	bigExpected, _ = big.NewInt(0).SetString(expected, 10)
	result = core.GetIntTrimmedPercentageOfValue(bigA, percentage)
	require.Equal(t, bigExpected, result)

	a = "29815853976407917651"
	percentage = 0.0000000001
	expected = "2981585397"
	bigA, _ = big.NewInt(0).SetString(a, 10)
	bigExpected, _ = big.NewInt(0).SetString(expected, 10)
	result = core.GetIntTrimmedPercentageOfValue(bigA, percentage)
	require.Equal(t, bigExpected, result)
}

func TestSplitExponentFraction(t *testing.T) {
	a := "123.01234567890123456789"
	expectedExp := "123"
	expectedFra := "01234567890123456789"
	exp, fra := core.SplitExponentFraction(a)
	require.Equal(t, expectedExp, exp)
	require.Equal(t, expectedFra, fra)

	a = "12345678901234567890"
	expectedExp = "12345678901234567890"
	expectedFra = ""
	exp, fra = core.SplitExponentFraction(a)
	require.Equal(t, expectedExp, exp)
	require.Equal(t, expectedFra, fra)

	a = "123.0123456789012345678900000000"
	expectedExp = "123"
	expectedFra = "0123456789012345678900000000"
	exp, fra = core.SplitExponentFraction(a)
	require.Equal(t, expectedExp, exp)
	require.Equal(t, expectedFra, fra)

	a = "00123.01234567890123456789000"
	expectedExp = "00123"
	expectedFra = "01234567890123456789000"
	exp, fra = core.SplitExponentFraction(a)
	require.Equal(t, expectedExp, exp)
	require.Equal(t, expectedFra, fra)

	a = "0.01234567890123456789000"
	expectedExp = "0"
	expectedFra = "01234567890123456789000"
	exp, fra = core.SplitExponentFraction(a)
	require.Equal(t, expectedExp, exp)
	require.Equal(t, expectedFra, fra)
}

func BenchmarkSplitExponentFraction(b *testing.B) {
	nbPrepared := 100000
	src := rand.NewSource(1122334455)
	r := rand.New(src)

	preparedFractionals := make([]string, nbPrepared)
	for i := 0; i < nbPrepared; i++ {
		f := r.Float64()

		preparedFractionals[i] = big.NewFloat(f).String()
	}

	for n := 0; n < b.N; n++ {
		core.SplitExponentFraction(preparedFractionals[n%nbPrepared])
	}
}

func BenchmarkGetIntTrimmedPercentageOfValue(b *testing.B) {
	nbPrepared := 100000
	src := rand.NewSource(1122334455)
	r := rand.New(src)

	preparedFractionals := make([]float64, nbPrepared)
	preparedBigInts := make([]*big.Int, nbPrepared)

	for i := 0; i < nbPrepared; i++ {
		f := r.Float64()
		preparedFractionals[i] = f

		factor := r.Uint64()
		seed, _ := big.NewInt(0).SetString("1123456789123456789", 10)
		preparedBigInts[i] = seed.Mul(seed, big.NewInt(0).SetUint64(factor))
	}

	for n := 0; n < b.N; n++ {
		core.GetIntTrimmedPercentageOfValue(preparedBigInts[n%nbPrepared], preparedFractionals[n%nbPrepared])
	}
}

func BenchmarkGetApproximatePercentageOfValue(b *testing.B) {
	nbPrepared := 100000
	src := rand.NewSource(1122334455)
	r := rand.New(src)

	preparedFractionals := make([]float64, nbPrepared)
	preparedBigInts := make([]*big.Int, nbPrepared)

	for i := 0; i < nbPrepared; i++ {
		f := r.Float64()
		preparedFractionals[i] = f

		factor := r.Uint64()
		seed, _ := big.NewInt(0).SetString("1123456789123456789", 10)
		preparedBigInts[i] = seed.Mul(seed, big.NewInt(0).SetUint64(factor))
	}

	for n := 0; n < b.N; n++ {
		core.GetApproximatePercentageOfValue(preparedBigInts[n%nbPrepared], preparedFractionals[n%nbPrepared])
	}
}
