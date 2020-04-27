package mcl

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPointGT(t *testing.T) {
	pGT := NewPointGT()
	require.NotNil(t, pGT)

	mclPointGT, ok := pGT.GetUnderlyingObj().(*bls.GT)
	require.True(t, ok)
	require.True(t, mclPointGT.IsZero())
}

func TestPointGT_Equal(t *testing.T) {
	p1GT := NewPointGT()
	p2GT := NewPointGT()

	eq, err := p1GT.Equal(p2GT)
	require.Nil(t, err)
	require.True(t, eq)

	p2GT.SetInt64(10)
	eq, err = p1GT.Equal(p2GT)
	require.Nil(t, err)
	require.False(t, eq)
}

func TestPointGT_CloneNilShouldPanic(t *testing.T) {
	var p1 *PointGT

	defer func() {
		r := recover()
		if r == nil {
			assert.Fail(t, "should have panicked")
		}
	}()

	_ = p1.Clone()
}

func TestPointGT_Clone(t *testing.T) {
	p1 := NewPointGT()
	p1.GT.SetInt64(10)
	p2 := p1.Clone()

	eq, err := p1.Equal(p2)
	require.Nil(t, err)
	require.True(t, eq)
}

func TestPointGT_Null(t *testing.T) {
	p1 := NewPointGT()

	point := p1.Null()
	mclPoint, ok := point.(*PointGT)
	require.True(t, ok)
	require.True(t, mclPoint.IsZero())
	mclPointNeg := &bls.GT{}
	bls.GTNeg(mclPointNeg, mclPoint.GT)

	// neutral identity point should be equal to it's negation
	require.True(t, mclPoint.IsEqual(mclPointNeg))
}

func TestPointGT_Set(t *testing.T) {
	p1 := NewPointGT()
	p2 := NewPointGT()

	p2.GT.SetInt64(10)

	err := p1.Set(p2)
	require.Nil(t, err)

	eq, err := p1.Equal(p2)
	require.Nil(t, err)
	require.True(t, eq)
}

func TestPointGT_AddNilParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointGT()
	point2, err := point.Add(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, point2)
}

func TestPointGT_AddInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointGT()
	point2 := &mock.PointMock{}
	point3, err := point.Add(point2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, point3)
}

func TestPointGT_AddOK(t *testing.T) {
	t.Parallel()

	pointGT := NewPointGT()
	point1, err := pointGT.Pick()
	require.Nil(t, err)

	point2, err := pointGT.Pick()
	require.Nil(t, err)

	sum, err := point1.Add(point2)
	require.Nil(t, err)

	p, err := sum.Sub(point2)
	require.Nil(t, err)

	eq1, _ := point1.Equal(sum)
	eq2, _ := point2.Equal(sum)
	eq3, _ := point1.Equal(p)

	assert.False(t, eq1)
	assert.False(t, eq2)
	assert.True(t, eq3)
}

func TestPointGT_SubNilParamShouldErr(t *testing.T) {
	t.Parallel()

	pointGT := NewPointGT()
	point2, err := pointGT.Sub(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, point2)
}

func TestPointGT_SubInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	pointGT := NewPointGT()
	point2 := &mock.PointMock{}
	point3, err := pointGT.Sub(point2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, point3)
}

func TestPointGT_SubOK(t *testing.T) {
	t.Parallel()

	pointGT := NewPointGT()
	point1, err := pointGT.Pick()
	require.Nil(t, err)

	point2, err := pointGT.Pick()
	require.Nil(t, err)

	sum, _ := point1.Add(point2)
	point3, err := sum.Sub(point2)
	assert.Nil(t, err)

	eq, err := point3.Equal(point1)
	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestPointGT_Neg(t *testing.T) {
	point1 := NewPointGT()
	point, _ := point1.Pick()

	point2 := point.Neg()
	point3 := point2.Neg()

	assert.NotEqual(t, point, point2)
	assert.NotEqual(t, point2, point3)
	assert.Equal(t, point, point3)
}

func TestPointGT_MulNilParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointGT()
	res, err := point.Mul(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, res)
}

func TestPointGT_MulInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointGT()
	scalar := &mock.ScalarMock{}
	res, err := point.Mul(scalar)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, res)
}

func TestPointGT_MulOK(t *testing.T) {
	t.Parallel()

	pointGT := NewPointGT()
	point2, _ := pointGT.Pick()
	s := NewScalar()
	scalar, err := s.Pick()
	require.Nil(t, err)

	res, err := point2.Mul(scalar)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.NotEqual(t, point2, res)
}

func TestPointGT_PickOK(t *testing.T) {
	t.Parallel()

	point1 := NewPointGT()
	point2, err1 := point1.Pick()
	eq, err2 := point1.Equal(point2)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.False(t, eq)
}

func TestPointGT_GetUnderlyingObj(t *testing.T) {
	t.Parallel()

	point1 := NewPointGT()
	p := point1.GetUnderlyingObj()

	assert.NotNil(t, p)
}

func TestPointGT_MarshalBinary(t *testing.T) {
	t.Parallel()

	point1 := NewPointGT()
	pointBytes, err := point1.MarshalBinary()

	assert.Nil(t, err)
	assert.NotNil(t, pointBytes)
}

func TestPointGT_UnmarshalBinary(t *testing.T) {
	t.Parallel()

	point1, _ := NewPointGT().Pick()
	pointBytes, _ := point1.MarshalBinary()

	point2 := NewPointGT()
	err := point2.UnmarshalBinary(pointBytes)
	eq, _ := point1.Equal(point2)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestPointGT_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var point *PointGT

	require.True(t, point.IsInterfaceNil())
	point = NewPointGT()
	require.False(t, point.IsInterfaceNil())
}
