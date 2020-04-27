package mcl

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPointG1Str = "1 2804441700383207137291059272726320202409298495939644656032276750978957042938306743947959295841222776376494045956033 549030927663217958779900832672186439800557853951422598612219049215978530222727658974829308838610188609159619602487"
const testPointG1StrBase = 10

func TestNewPointG1(t *testing.T) {
	pG1 := NewPointG1()
	require.NotNil(t, pG1)

	bG1 := &bls.G1{}
	baseG1Str := baseG1()
	err := bG1.SetString(baseG1Str, 10)
	require.Nil(t, err)

	mclPointG1, ok := pG1.GetUnderlyingObj().(*bls.G1)
	require.True(t, ok)
	require.True(t, mclPointG1.IsValid())
	require.True(t, mclPointG1.IsValidOrder())
	require.False(t, mclPointG1.IsZero())
	require.True(t, bG1.IsEqual(mclPointG1))
}

func TestPointG1_Equal(t *testing.T) {
	p1G1 := NewPointG1()
	p2G1 := NewPointG1()

	// new points should be initialized with base point so should be equal
	eq, err := p1G1.Equal(p2G1)
	require.Nil(t, err)
	require.True(t, eq)

	err = p1G1.SetString(testPointG1Str, testPointG1StrBase)
	require.Nil(t, err)

	eq, err = p1G1.Equal(p2G1)
	require.Nil(t, err)
	require.False(t, eq)

	grG1 := &groupG1{}
	sc1G1 := grG1.CreateScalar()
	p1 := grG1.CreatePointForScalar(sc1G1)
	p2 := grG1.CreatePointForScalar(sc1G1)

	var ok bool
	p1G1, ok = p1.(*PointG1)
	require.True(t, ok)

	p2G1, ok = p2.(*PointG1)
	require.True(t, ok)

	eq, err = p1G1.Equal(p2G1)
	require.Nil(t, err)
	require.True(t, eq)
}

func TestPointG1_CloneNilShouldPanic(t *testing.T) {
	var p1 *PointG1

	defer func() {
		r := recover()
		if r == nil {
			assert.Fail(t, "should have panicked")
		}
	}()

	_ = p1.Clone()
}

func TestPointG1_Clone(t *testing.T) {
	p1 := NewPointG1()
	p2 := p1.Clone()

	eq, err := p1.Equal(p2)
	require.Nil(t, err)
	require.True(t, eq)
}

func TestPointG1_Null(t *testing.T) {
	p1 := NewPointG1()

	point := p1.Null()
	mclPoint, ok := point.(*PointG1)
	require.True(t, ok)
	require.True(t, mclPoint.IsZero())
	mclPointNeg := &bls.G1{}
	bls.G1Neg(mclPointNeg, mclPoint.G1)

	// neutral identity point should be equal to it's negation
	require.True(t, mclPoint.IsEqual(mclPointNeg))
}

func TestPointG1_Set(t *testing.T) {
	p1 := NewPointG1()
	p2 := NewPointG1()

	err := p2.SetString(testPointG1Str, testPointG1StrBase)
	require.Nil(t, err)

	err = p1.Set(p2)
	require.Nil(t, err)
	eq, err := p1.Equal(p2)
	require.Nil(t, err)
	require.True(t, eq)
}

func TestPointG1_AddNilParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointG1()
	point2, err := point.Add(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, point2)
}

func TestPointG1_AddInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointG1()
	point2 := &mock.PointMock{}
	point3, err := point.Add(point2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, point3)
}

func TestPointG1_AddOK(t *testing.T) {
	t.Parallel()

	pointG1 := NewPointG1()
	point1, err := pointG1.Pick()
	require.Nil(t, err)

	point2, err := pointG1.Pick()
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

func TestPointG1_SubNilParamShouldErr(t *testing.T) {
	t.Parallel()

	pointG1 := NewPointG1()
	point2, err := pointG1.Sub(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, point2)
}

func TestPointG1_SubInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	pointG1 := NewPointG1()
	point2 := &mock.PointMock{}
	point3, err := pointG1.Sub(point2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, point3)
}

func TestPointG1_SubOK(t *testing.T) {
	t.Parallel()

	pointG1 := NewPointG1()
	point1, err := pointG1.Pick()
	require.Nil(t, err)

	point2, err := pointG1.Pick()
	require.Nil(t, err)

	sum, _ := point1.Add(point2)
	point3, err := sum.Sub(point2)
	assert.Nil(t, err)

	eq, err := point3.Equal(point1)
	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestPointG1_Neg(t *testing.T) {
	point1 := NewPointG1()

	point2 := point1.Neg()
	point3 := point2.Neg()

	assert.NotEqual(t, point1, point2)
	assert.NotEqual(t, point2, point3)
	assert.Equal(t, point1, point3)
}

func TestPointG1_MulNilParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointG1()
	res, err := point.Mul(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, res)
}

func TestPointG1_MulInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	point := NewPointG1()
	scalar := &mock.ScalarMock{}
	res, err := point.Mul(scalar)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, res)
}

func TestPointG1_MulOK(t *testing.T) {
	t.Parallel()

	pointG1 := NewPointG1()
	s := NewScalar()
	scalar, err := s.Pick()
	require.Nil(t, err)

	res, err := pointG1.Mul(scalar)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.NotEqual(t, pointG1, res)

	grG1 := &groupG1{}
	point2 := grG1.CreatePointForScalar(scalar)
	eq, err := res.Equal(point2)
	require.Nil(t, err)
	require.True(t, eq)
}

func TestPointG1_PickOK(t *testing.T) {
	t.Parallel()

	point1 := NewPointG1()
	point2, err1 := point1.Pick()
	eq, err2 := point1.Equal(point2)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.False(t, eq)
}

func TestPointG1_GetUnderlyingObj(t *testing.T) {
	t.Parallel()

	point1 := NewPointG1()
	p := point1.GetUnderlyingObj()

	assert.NotNil(t, p)
}

func TestPointG1_MarshalBinary(t *testing.T) {
	t.Parallel()

	point1 := NewPointG1()
	pointBytes, err := point1.MarshalBinary()

	assert.Nil(t, err)
	assert.NotNil(t, pointBytes)
}

func TestPointG1_UnmarshalBinary(t *testing.T) {
	t.Parallel()

	point1, _ := NewPointG1().Pick()
	pointBytes, _ := point1.MarshalBinary()

	point2 := NewPointG1()
	err := point2.UnmarshalBinary(pointBytes)
	eq, _ := point1.Equal(point2)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestPointG1_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var point *PointG1

	require.True(t, check.IfNil(point))
	point = NewPointG1()
	require.False(t, check.IfNil(point))
}
