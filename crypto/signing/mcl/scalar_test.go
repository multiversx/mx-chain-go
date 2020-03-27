package mcl

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/stretchr/testify/require"
)

func TestMclScalar_EqualNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar := suite.CreateScalar().Zero()

	eq, err := scalar.Equal(nil)

	require.False(t, eq)
	require.Equal(t, crypto.ErrNilParam, err)
}

func TestMclScalar_EqualInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := &mock.ScalarMock{}
	eq, err := scalar1.Equal(scalar2)

	require.False(t, eq)
	require.Equal(t, crypto.ErrInvalidParam, err)
}

func TestMclScalar_EqualTrue(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().One()
	eq, err := scalar1.Equal(scalar2)

	require.Nil(t, err)
	require.True(t, eq)
}

func TestMclScalar_EqualFalse(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().Zero()
	eq, err := scalar1.Equal(scalar2)

	require.Nil(t, err)
	require.False(t, eq)
}

func TestMclScalar_SetNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar := suite.CreateScalar().One()
	err := scalar.Set(nil)

	require.Equal(t, crypto.ErrNilParam, err)
}

func TestMclScalar_SetInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := &mock.ScalarMock{}
	err := scalar1.Set(scalar2)

	require.Equal(t, crypto.ErrInvalidParam, err)
}

func TestMclScalar_SetOK(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().Zero()
	err := scalar1.Set(scalar2)
	eq, _ := scalar1.Equal(scalar2)

	require.Nil(t, err)
	require.True(t, eq)
}

func TestMclScalar_Clone(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := scalar1.Clone()
	eq, err := scalar1.Equal(scalar2)

	require.Nil(t, err)
	require.True(t, eq)
}

func TestMclScalar_SetInt64(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar2 := suite.CreateScalar()
	scalar1.SetInt64(int64(555555555))
	scalar2.SetInt64(int64(444444444))

	diff, _ := scalar1.Sub(scalar2)
	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(int64(111111111))

	eq, err := diff.Equal(scalar3)

	require.Nil(t, err)
	require.True(t, eq)
}

func TestMclScalar_Zero(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := suite.CreateScalar()
	scalar2.SetInt64(0)

	eq, err := scalar2.Equal(scalar1)

	require.Nil(t, err)
	require.True(t, eq)
}

func TestMclScalar_AddNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar := suite.CreateScalar().Zero()
	sum, err := scalar.Add(nil)

	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, sum)
}

func TestMclScalar_AddInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := &mock.ScalarMock{}
	sum, err := scalar1.Add(scalar2)

	require.Equal(t, crypto.ErrInvalidParam, err)
	require.Nil(t, sum)
}

func TestMclScalar_AddOK(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().One()
	sum, err := scalar1.Add(scalar2)
	require.Nil(t, err)
	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(2)
	eq, err := scalar3.Equal(sum)

	require.True(t, eq)
	require.Nil(t, err)
}

func TestMclScalar_SubNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar := suite.CreateScalar().Zero()
	diff, err := scalar.Sub(nil)

	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, diff)
}

func TestMclScalar_SubInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := &mock.ScalarMock{}
	diff, err := scalar1.Sub(scalar2)

	require.Equal(t, crypto.ErrInvalidParam, err)
	require.Nil(t, diff)
}

func TestMclScalar_SubOK(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(4)
	scalar2 := suite.CreateScalar().One()
	diff, err := scalar1.Sub(scalar2)
	require.Nil(t, err)
	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(3)
	eq, err := scalar3.Equal(diff)

	require.True(t, eq)
	require.Nil(t, err)
}

func TestMclScalar_Neg(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(4)
	scalar2 := scalar1.Neg()
	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(-4)
	eq, err := scalar2.Equal(scalar3)

	require.Nil(t, err)
	require.True(t, eq)
}

func TestMclScalar_One(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(1)
	scalar2 := suite.CreateScalar().One()

	eq, err := scalar1.Equal(scalar2)

	require.Nil(t, err)
	require.True(t, eq)
}

func TestMclScalar_MulNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar := suite.CreateScalar().One()
	res, err := scalar.Mul(nil)

	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, res)
}

func TestMclScalar_MulInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := &mock.ScalarMock{}
	res, err := scalar1.Mul(scalar2)

	require.Equal(t, crypto.ErrInvalidParam, err)
	require.Nil(t, res)
}

func TestMclScalar_MulOK(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar()
	scalar2.SetInt64(4)
	res, err := scalar1.Mul(scalar2)

	require.Nil(t, err)

	eq, _ := res.Equal(scalar2)

	require.True(t, eq)
}

func TestMclScalar_DivNilParamShouldEr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar := suite.CreateScalar().One()
	res, err := scalar.Div(nil)

	require.Equal(t, crypto.ErrNilParam, err)
	require.Nil(t, res)
}

func TestMclScalar_DivInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := &mock.ScalarMock{}
	res, err := scalar1.Div(scalar2)

	require.Equal(t, crypto.ErrInvalidParam, err)
	require.Nil(t, res)
}

func TestMclScalar_DivOK(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar()
	scalar2.SetInt64(4)
	res, err := scalar2.Div(scalar1)

	require.Nil(t, err)

	eq, _ := res.Equal(scalar2)

	require.True(t, eq)
}

func TestMclScalar_InvNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar2, err := scalar1.Inv(nil)

	require.Nil(t, scalar2)
	require.Equal(t, crypto.ErrNilParam, err)
}

func TestMclScalar_InvInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar2 := &mock.ScalarMock{}
	scalar3, err := scalar1.Inv(scalar2)

	require.Nil(t, scalar3)
	require.Equal(t, crypto.ErrInvalidParam, err)
}

func TestMclScalar_InvOK(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(4)
	scalar2, err := scalar1.Inv(scalar1)
	eq, _ := scalar1.Equal(scalar2)

	require.Nil(t, err)
	require.NotNil(t, scalar2)
	require.False(t, eq)

	one := suite.CreateScalar().One()
	scalar1, err = scalar1.Inv(one)
	require.Nil(t, err)
	eq, _ = one.Equal(scalar1)

	require.True(t, eq)
}

func TestMclScalar_PickOK(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar2, err := scalar1.Pick()
	require.Nil(t, err)
	require.NotNil(t, scalar1, scalar2)

	eq, _ := scalar1.Equal(scalar2)

	require.False(t, eq)
}

func TestMclScalar_SetBytesNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar()
	scalar2, err := scalar1.SetBytes(nil)

	require.Nil(t, scalar2)
	require.Equal(t, crypto.ErrNilParam, err)
}

func TestMclScalar_SetBytesOK(t *testing.T) {
	t.Parallel()

	val := int64(555555555)
	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()

	sc2 := NewScalar()
	sc2.SetInt64(val)
	buf, _ := sc2.MarshalBinary()

	scalar2, err := scalar1.SetBytes(buf)
	require.Nil(t, err)
	require.NotEqual(t, scalar1, scalar2)

	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(val)

	eq, _ := scalar3.Equal(scalar2)
	require.True(t, eq)
}

func TestMclScalar_GetUnderlyingObj(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()
	x := scalar1.GetUnderlyingObj()

	require.NotNil(t, x)
}

func TestMclScalar_MarshalBinary(t *testing.T) {
	t.Parallel()

	suite := NewSuiteBLS12()
	scalar1 := suite.CreateScalar().One()

	scalarBytes, err := scalar1.MarshalBinary()

	require.Nil(t, err)
	require.NotNil(t, scalarBytes)
}

func TestMclScalar_UnmarshalBinary(t *testing.T) {
	suite := NewSuiteBLS12()
	scalar1, _ := suite.CreateScalar().Pick()
	scalarBytes, err := scalar1.MarshalBinary()
	require.Nil(t, err)
	scalar2 := suite.CreateScalar().Zero()
	err = scalar2.UnmarshalBinary(scalarBytes)
	require.Nil(t, err)

	eq, err := scalar1.Equal(scalar2)

	require.Nil(t, err)
	require.True(t, eq)
}
