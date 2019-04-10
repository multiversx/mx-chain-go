package kyber_test

import (
	"encoding/binary"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
	"github.com/stretchr/testify/assert"
)

func TestKyberScalar_EqualNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar := suite.CreateScalar().Zero()

	eq, err := scalar.Equal(nil)

	assert.False(t, eq)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberScalar_EqualInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := &mock.ScalarMock{}
	eq, err := scalar1.Equal(scalar2)

	assert.False(t, eq)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKyberScalar_EqualTrue(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().One()
	eq, err := scalar1.Equal(scalar2)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberScalar_EqualFalse(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().Zero()
	eq, err := scalar1.Equal(scalar2)

	assert.Nil(t, err)
	assert.False(t, eq)
}

func TestKyberScalar_SetNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar := suite.CreateScalar().One()
	err := scalar.Set(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberScalar_SetInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := &mock.ScalarMock{}
	err := scalar1.Set(scalar2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKyberScalar_SetOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().Zero()
	err := scalar1.Set(scalar2)
	eq, _ := scalar1.Equal(scalar2)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberScalar_Clone(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := scalar1.Clone()
	eq, err := scalar1.Equal(scalar2)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberScalar_SetInt64(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar2 := suite.CreateScalar()
	scalar1.SetInt64(int64(555555555))
	scalar2.SetInt64(int64(444444444))

	diff, _ := scalar1.Sub(scalar2)
	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(int64(111111111))

	eq, err := diff.Equal(scalar3)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberScalar_Zero(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := suite.CreateScalar()
	scalar2.SetInt64(0)

	eq, err := scalar2.Equal(scalar1)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberScalar_AddNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar := suite.CreateScalar().Zero()
	sum, err := scalar.Add(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, sum)
}

func TestKyberScalar_AddInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := &mock.ScalarMock{}
	sum, err := scalar1.Add(scalar2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, sum)
}

func TestKyberScalar_AddOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar().One()
	sum, err := scalar1.Add(scalar2)

	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(2)
	eq, err := scalar3.Equal(sum)

	assert.True(t, eq)
	assert.Nil(t, err)
}

func TestKyberScalar_SubNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar := suite.CreateScalar().Zero()
	diff, err := scalar.Sub(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, diff)
}

func TestKyberScalar_SubInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().Zero()
	scalar2 := &mock.ScalarMock{}
	diff, err := scalar1.Sub(scalar2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, diff)
}

func TestKyberScalar_SubOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(4)
	scalar2 := suite.CreateScalar().One()
	diff, err := scalar1.Sub(scalar2)

	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(3)
	eq, err := scalar3.Equal(diff)

	assert.True(t, eq)
	assert.Nil(t, err)
}

func TestKyberScalar_Neg(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(4)
	scalar2 := scalar1.Neg()
	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(-4)
	eq, err := scalar2.Equal(scalar3)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberScalar_One(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(1)
	scalar2 := suite.CreateScalar().One()

	eq, err := scalar1.Equal(scalar2)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberScalar_MulNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar := suite.CreateScalar().One()
	res, err := scalar.Mul(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, res)
}

func TestKyberScalar_MulInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := &mock.ScalarMock{}
	res, err := scalar1.Mul(scalar2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, res)
}

func TestKyberScalar_MulOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar()
	scalar2.SetInt64(4)
	res, err := scalar1.Mul(scalar2)

	assert.Nil(t, err)

	eq, _ := res.Equal(scalar2)

	assert.True(t, eq)
}

func TestKyberScalar_DivNilParamShouldEr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar := suite.CreateScalar().One()
	res, err := scalar.Div(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, res)
}

func TestKyberScalar_DivInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := &mock.ScalarMock{}
	res, err := scalar1.Div(scalar2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, res)
}

func TestKyberScalar_DivOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	scalar2 := suite.CreateScalar()
	scalar2.SetInt64(4)
	res, err := scalar2.Div(scalar1)

	assert.Nil(t, err)

	eq, _ := res.Equal(scalar2)

	assert.True(t, eq)
}

func TestKyberScalar_InvNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar2, err := scalar1.Inv(nil)

	assert.Nil(t, scalar2)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberScalar_InvInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar2 := &mock.ScalarMock{}
	scalar3, err := scalar1.Inv(scalar2)

	assert.Nil(t, scalar3)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKyberScalar_InvOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar1.SetInt64(4)
	scalar2, err := scalar1.Inv(scalar1)
	eq, _ := scalar1.Equal(scalar2)

	assert.Nil(t, err)
	assert.NotNil(t, scalar2)
	assert.False(t, eq)

	one := suite.CreateScalar().One()
	scalar1, err = scalar1.Inv(one)
	eq, _ = one.Equal(scalar1)

	assert.True(t, eq)
}

func TestKyberScalar_PickNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()

	scalar2, err := scalar1.Pick(nil)

	assert.Nil(t, scalar2)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberScalar_PickOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	scalar1 := suite.CreateScalar()
	scalar2, err := scalar1.Pick(stream)
	assert.Nil(t, err)
	assert.NotNil(t, scalar1, scalar2)

	eq, _ := scalar1.Equal(scalar2)

	assert.False(t, eq)
}

func TestKyberScalar_SetBytesNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar()
	scalar2, err := scalar1.SetBytes(nil)

	assert.Nil(t, scalar2)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberScalar_SetBytesOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	buf := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(buf, 555555555)

	scalar2, err := scalar1.SetBytes(buf)
	assert.Nil(t, err)
	assert.NotEqual(t, scalar1, scalar2)

	scalar3 := suite.CreateScalar()
	scalar3.SetInt64(555555555)

	eq, _ := scalar3.Equal(scalar2)
	assert.True(t, eq)
}

func TestKyberScalar_GetUnderlyingObj(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()
	x := scalar1.GetUnderlyingObj()

	assert.NotNil(t, x)
}

func TestKyberScalar_MarshalBinary(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	scalar1 := suite.CreateScalar().One()

	scalarBytes, err := scalar1.MarshalBinary()

	assert.Nil(t, err)
	assert.NotNil(t, scalarBytes)
}

func TestKyberScalar_UnmarshalBinary(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	randStream := suite.RandomStream()
	scalar1, _ := suite.CreateScalar().Pick(randStream)
	scalarBytes, err := scalar1.MarshalBinary()

	scalar2 := suite.CreateScalar().Zero()
	scalar2.UnmarshalBinary(scalarBytes)

	eq, err := scalar1.Equal(scalar2)

	assert.Nil(t, err)
	assert.True(t, eq)
}
