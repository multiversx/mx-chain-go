package mcl

import (
	"testing"

	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/require"
)

func TestGroupG1_String(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}

	str := grG1.String()
	require.Equal(t, str, "BLS12-381 G1")
}

func TestGroupG1_ScalarLen(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}

	x := grG1.ScalarLen()
	require.Equal(t, 32, x)
}

func TestGroupG1_PointLen(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}

	x := grG1.PointLen()
	require.Equal(t, 48, x)
}

func TestGroupG1_CreatePoint(t *testing.T) {
	t.Parallel()

	baseG1Str := baseG1()

	grG1 := &groupG1{}
	point := &PointG1{
		G1: &bls.G1{},
	}

	err := point.G1.SetString(baseG1Str, 10)
	require.Nil(t, err)
	require.False(t, point.G1.IsZero())
	require.True(t, point.G1.IsValid())
	require.True(t, point.G1.IsValidOrder())

	x := grG1.CreatePoint()
	require.NotNil(t, x)
	mclPoint, ok := x.GetUnderlyingObj().(*bls.G1)
	require.True(t, ok)
	require.False(t, mclPoint.IsZero())
	require.True(t, mclPoint.IsValid())
	require.True(t, mclPoint.IsValidOrder())
	require.Equal(t, point.G1, mclPoint)
}

func TestGroupG1_CreateScalar(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}

	sc := grG1.CreateScalar()
	require.NotNil(t, sc)

	mclScalar, ok := sc.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.False(t, mclScalar.IsZero())
	require.True(t, mclScalar.IsValid())
	require.False(t, mclScalar.IsOne())
}

func TestGroupG1_CreatePointForScalar(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}

	scalar := grG1.CreateScalar()
	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.False(t, mclScalar.IsZero())
	require.False(t, mclScalar.IsOne())
	require.True(t, mclScalar.IsValid())

	pG1 := grG1.CreatePointForScalar(scalar)
	require.NotNil(t, pG1)

	mclPointG1, ok := pG1.GetUnderlyingObj().(*bls.G1)
	require.True(t, ok)
	require.False(t, mclPointG1.IsZero())
	require.True(t, mclPointG1.IsValidOrder())
	require.True(t, mclPointG1.IsValid())

	baseG1 := NewPointG1().G1
	computedG1 := &bls.G1{}

	bls.G1Mul(computedG1, baseG1, mclScalar)
	require.True(t, mclPointG1.IsEqual(computedG1))
}

func TestGroupG1_CreatePointForScalarZero(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}

	scalar := grG1.CreateScalar()
	scalar.SetInt64(0)
	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.True(t, mclScalar.IsZero())
	require.True(t, mclScalar.IsValid())

	pG1 := grG1.CreatePointForScalar(scalar)
	require.NotNil(t, pG1)

	mclPointG1, ok := pG1.GetUnderlyingObj().(*bls.G1)
	require.True(t, ok)
	require.True(t, mclPointG1.IsZero())
	require.True(t, mclPointG1.IsValidOrder())
	require.True(t, mclPointG1.IsValid())

	baseG1 := NewPointG1().G1
	computedG1 := &bls.G1{}

	bls.G1Mul(computedG1, baseG1, mclScalar)
	require.True(t, mclPointG1.IsEqual(computedG1))
}

func TestGroupG1_CreatePointForScalarOne(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}

	scalar := grG1.CreateScalar()
	scalar.SetInt64(1)
	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.True(t, mclScalar.IsOne())
	require.True(t, mclScalar.IsValid())

	pG1 := grG1.CreatePointForScalar(scalar)
	require.NotNil(t, pG1)

	baseG1 := NewPointG1().G1
	mclPointG1, ok := pG1.GetUnderlyingObj().(*bls.G1)
	require.True(t, ok)
	require.True(t, mclPointG1.IsEqual(baseG1))
	require.True(t, mclPointG1.IsValidOrder())
	require.True(t, mclPointG1.IsValid())
}

func TestGroupG1_CreatePointForScalarNil(t *testing.T) {
	t.Parallel()

	grG1 := &groupG1{}
	pG1 := grG1.CreatePointForScalar(nil)
	require.Equal(t, nil, pG1)
}

func TestGroupG1_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var grG1 *groupG1

	require.True(t, grG1.IsInterfaceNil())
	grG1 = &groupG1{}
	require.False(t, grG1.IsInterfaceNil())
}
