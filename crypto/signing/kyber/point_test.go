package kyber_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/stretchr/testify/assert"
)

func TestKyberPoint_EqualShouldNotChangeReceiver(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	point2 := suite.CreatePoint().Null()
	pointClone := point.Clone()
	eq, err := point.Equal(point2)

	assert.NotEqual(t, point, point2)
	assert.False(t, eq)
	assert.Nil(t, err)
	assert.Equal(t, pointClone, point)
}

func TestKyberPoint_EqualNilParamShouldRetError(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	eq, err := point.Equal(nil)

	assert.False(t, eq)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberPoint_EqualInvalidParamShouldRetErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	point2 := &mock.PointMock{}
	eq, err := point.Equal(point2)

	assert.False(t, eq)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKyberPoint_EqualShouldRetFalse(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	point2 := suite.CreatePoint().Base().Neg()
	eq, err := point.Equal(point2)

	assert.Nil(t, err)
	assert.False(t, eq)
}

func TestKyberPoint_EqualShouldRetTrue(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	point2 := suite.CreatePoint().Base()
	eq, err := point.Equal(point2)

	assert.Nil(t, err)
	assert.True(t, eq)
}

func TestKyberPoint_NullShouldNotChangeReceiver(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	point2 := point.Clone()
	point3 := point.Null()

	assert.NotEqual(t, point2, point3)
	assert.Equal(t, point, point2)
}

func TestKyberPoint_NullOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	point2 := point.Null()
	point3 := point2.Neg()

	assert.NotEqual(t, point, point2)
	// neutral identity point should be equal to it's negation
	assert.Equal(t, point2, point3)
}

func TestKyberPoint_BaseShouldNotChangeReceiver(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	point, _ := suite.CreatePoint().Pick(stream)
	point2 := point.Clone()
	point3 := point.Base()

	assert.Equal(t, point, point2)
	assert.NotEqual(t, point, point3)
}

func TestKyberPoint_BaseOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	assert.NotNil(t, point)
}

func TestKyberPoint_SetNilParamShouldRetErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()

	err := point.Set(nil)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberPoint_SetInvalidParamShouldRetErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint().Base()
	point2 := &mock.PointMock{}
	err := point.Set(point2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKyberPoint_SetOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	point := suite.CreatePoint().Base()
	point2, _ := suite.CreatePoint().Pick(stream)
	err := point.Set(point2)

	assert.Nil(t, err)
	assert.Equal(t, point, point2)
}

func TestKyberPoint_CloneShouldNotChangeReceiver(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	point2 := point.Clone()

	assert.False(t, point == point2)
}

func TestKyberPoint_CloneOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	point2 := point.Clone()

	assert.Equal(t, point, point2)
}

func TestKyberPoint_AddNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	point2, err := point.Add(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, point2)
}

func TestKyberPoint_AddInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	point2 := &mock.PointMock{}
	point3, err := point.Add(point2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, point3)
}

func TestKyberPoint_AddOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	point1 := suite.CreatePoint()
	point2, _ := suite.CreatePoint().Pick(stream)

	sum, err := point1.Add(point2)
	assert.Nil(t, err)

	p, _ := sum.Sub(point2)

	eq1, _ := point1.Equal(sum)
	eq2, _ := point2.Equal(sum)
	eq3, _ := point1.Equal(p)

	assert.False(t, eq1)
	assert.False(t, eq2)
	assert.True(t, eq3)
}

func TestKyberPoint_SubNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	point2, err := point.Sub(nil)

	assert.Equal(t, crypto.ErrNilParam, err)
	assert.Nil(t, point2)
}

func TestKyberPoint_SubInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	point2 := &mock.PointMock{}
	point3, err := point.Sub(point2)

	assert.Equal(t, crypto.ErrInvalidParam, err)
	assert.Nil(t, point3)
}

func TestKyberPoint_SubOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	point1 := suite.CreatePoint()
	point2, _ := suite.CreatePoint().Pick(stream)

	sum, _ := point1.Add(point2)
	point3, err := sum.Sub(point2)
	assert.Nil(t, err)

	eq, _ := point3.Equal(point1)

	assert.True(t, eq)
}

func TestKyberPoint_NegShouldNotChangeReceiver(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	pointClone := point.Clone()
	point2 := point.Neg()

	assert.NotEqual(t, point, point2)
	assert.Equal(t, point, pointClone)
}

func TestKyberPoint_NegOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	point2 := point.Neg()
	point3 := point2.Neg()

	assert.NotEqual(t, point, point2)
	assert.Equal(t, point, point3)
}

func TestKyberPoint_MulNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	res, err := point.Mul(nil)

	assert.Nil(t, res)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberPoint_MulInvalidParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point := suite.CreatePoint()
	scalar := &mock.ScalarMock{}
	res, err := point.Mul(scalar)

	assert.Nil(t, res)
	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestKyberPoint_MulOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	point := suite.CreatePoint()
	scalar, _ := suite.CreateScalar().Pick(stream)
	res, err := point.Mul(scalar)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEqual(t, point, res)
}

func TestKyberPoint_PickNilParamShouldErr(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	point1 := suite.CreatePoint()
	point2, err := point1.Pick(nil)

	assert.Nil(t, point2)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestKyberPoint_PickOK(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	point1 := suite.CreatePoint()
	point2, err1 := point1.Pick(stream)
	eq, err2 := point1.Equal(point2)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.False(t, eq)
}

func TestKyberPoint_GetUnderlyingObj(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	point1 := suite.CreatePoint().Base()
	p := point1.GetUnderlyingObj()

	assert.NotNil(t, p)
}

func TestKyberPoint_MarshalBinary(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	point1 := suite.CreatePoint().Base()
	pointBytes, err := point1.MarshalBinary()

	assert.Nil(t, err)
	assert.NotNil(t, pointBytes)
}

func TestKyberPoint_UnmarshalBinary(t *testing.T) {
	t.Parallel()

	suite := kyber.NewBlakeSHA256Ed25519()
	stream := suite.RandomStream()
	point1, _ := suite.CreatePoint().Pick(stream)
	pointBytes, _ := point1.MarshalBinary()

	point2 := suite.CreatePoint()
	err := point2.UnmarshalBinary(pointBytes)
	eq, _ := point1.Equal(point2)

	assert.Nil(t, err)
	assert.True(t, eq)
}
