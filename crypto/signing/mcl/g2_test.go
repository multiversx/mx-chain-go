package mcl

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/require"
)

func TestGroupG2_String(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}

	str := grG2.String()
	require.Equal(t, str, "BLS12-381 G2")
}

func TestGroupG2_ScalarLen(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}

	x := grG2.ScalarLen()
	require.Equal(t, 32, x)
}

func TestGroupG2_PointLen(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}

	x := grG2.PointLen()
	require.Equal(t, 96, x)
}

func TestGroupG2_CreatePoint(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}
	point := &PointG2{
		G2: &bls.G2{},
	}

	baseG2Str := baseG2()
	err := point.G2.SetString(baseG2Str, 10)
	require.Nil(t, err)
	require.False(t, point.G2.IsZero())
	require.True(t, point.G2.IsValid())
	require.True(t, point.G2.IsValidOrder())

	x := grG2.CreatePoint()
	require.NotNil(t, x)
	mclPoint, ok := x.GetUnderlyingObj().(*bls.G2)
	require.True(t, ok)
	require.False(t, mclPoint.IsZero())
	require.True(t, mclPoint.IsValid())
	require.True(t, mclPoint.IsValidOrder())
	require.Equal(t, point.G2, mclPoint)
}

func TestGroupG2_CreateScalar(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}

	sc := grG2.CreateScalar()
	require.NotNil(t, sc)

	mclScalar, ok := sc.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.False(t, mclScalar.IsZero())
	require.True(t, mclScalar.IsValid())
	require.False(t, mclScalar.IsOne())
}

func TestGroupG2_CreatePointForScalar(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}

	scalar := grG2.CreateScalar()
	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.False(t, mclScalar.IsZero())
	require.False(t, mclScalar.IsOne())
	require.True(t, mclScalar.IsValid())

	pG2 := grG2.CreatePointForScalar(scalar)
	require.NotNil(t, pG2)

	mclPointG2, ok := pG2.GetUnderlyingObj().(*bls.G2)
	require.True(t, ok)
	require.False(t, mclPointG2.IsZero())
	require.True(t, mclPointG2.IsValidOrder())
	require.True(t, mclPointG2.IsValid())

	bG2 := NewPointG2().G2
	computedG2 := &bls.G2{}

	bls.G2Mul(computedG2, bG2, mclScalar)
	require.True(t, mclPointG2.IsEqual(computedG2))
}

func TestGroupG2_CreatePointForScalarZero(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}

	scalar := grG2.CreateScalar()
	scalar.SetInt64(0)
	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.True(t, mclScalar.IsZero())
	require.True(t, mclScalar.IsValid())

	pG2 := grG2.CreatePointForScalar(scalar)
	require.NotNil(t, pG2)

	mclPointG2, ok := pG2.GetUnderlyingObj().(*bls.G2)
	require.True(t, ok)
	require.True(t, mclPointG2.IsZero())
	require.True(t, mclPointG2.IsValidOrder())
	require.True(t, mclPointG2.IsValid())

	bG2 := NewPointG2().G2
	computedG2 := &bls.G2{}

	bls.G2Mul(computedG2, bG2, mclScalar)
	require.True(t, mclPointG2.IsEqual(computedG2))
}

func TestGroupG2_CreatePointForScalarOne(t *testing.T) {
	t.Parallel()

	grG2 := &groupG2{}

	scalar := grG2.CreateScalar()
	scalar.SetInt64(1)
	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.True(t, mclScalar.IsOne())
	require.True(t, mclScalar.IsValid())

	pG2 := grG2.CreatePointForScalar(scalar)
	require.NotNil(t, pG2)

	bG2 := NewPointG2().G2
	mclPointG2, ok := pG2.GetUnderlyingObj().(*bls.G2)
	require.True(t, ok)
	require.True(t, mclPointG2.IsEqual(bG2))
	require.True(t, mclPointG2.IsValidOrder())
	require.True(t, mclPointG2.IsValid())
}

func TestGroupG2_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var grG2 *groupG2

	require.True(t, check.IfNil(grG2))
	grG2 = &groupG2{}
	require.False(t, check.IfNil(grG2))
}
