package mcl

import (
	"testing"

	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupGT_String(t *testing.T) {
	t.Parallel()

	grGT := &groupGT{}

	str := grGT.String()
	require.Equal(t, str, "BLS12-381 GT")
}

func TestGroupGT_ScalarLen(t *testing.T) {
	t.Parallel()

	grGT := &groupGT{}

	x := grGT.ScalarLen()
	require.Equal(t, 32, x)
}

func TestGroupGT_PointLen(t *testing.T) {
	t.Parallel()

	grGT := &groupGT{}

	x := grGT.PointLen()
	require.Equal(t, 48*12, x)
}

func TestGroupGT_CreatePoint(t *testing.T) {
	t.Parallel()

	grGT := &groupGT{}

	x := grGT.CreatePoint()
	require.NotNil(t, x)

	mclPoint, ok := x.GetUnderlyingObj().(*bls.GT)
	require.True(t, ok)
	// points created on GT are initialized with PointZero
	require.True(t, mclPoint.IsZero())
}

func TestGroupGT_CreateScalar(t *testing.T) {
	t.Parallel()

	grGT := &groupGT{}

	sc := grGT.CreateScalar()
	require.NotNil(t, sc)

	mclScalar, ok := sc.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.False(t, mclScalar.IsZero())
	require.True(t, mclScalar.IsValid())
	require.False(t, mclScalar.IsOne())
}

func TestGroupGT_CreatePointForScalar(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			assert.Fail(t, "should panic as currently not supported")
		}
	}()

	grGT := &groupGT{}

	scalar := grGT.CreateScalar()
	mclScalar, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	require.True(t, ok)
	require.False(t, mclScalar.IsZero())
	require.False(t, mclScalar.IsOne())
	require.True(t, mclScalar.IsValid())

	_ = grGT.CreatePointForScalar(scalar)
}

func TestGroupGT_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var grGT *groupGT

	require.True(t, grGT.IsInterfaceNil())
	grGT = &groupGT{}
	require.False(t, grGT.IsInterfaceNil())
}
