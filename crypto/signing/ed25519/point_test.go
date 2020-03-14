package ed25519_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestEd25519PointEqual_NilParamShouldErr(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()

	_, err := point.Equal(nil)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestEd25519PointEqual_InvalidParamShouldErr(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()

	point2 := &mock.PointMock{}
	_, err := point.Equal(point2)
	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestEd25519PointEqual_ReturnsTrueForTheSameKey(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()

	eq, _ := point.Equal(point)
	assert.True(t, eq)
}

func TestEd25519PointEqual_ReturnsFalseForDifferentKeys(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()
	point2 := suite.CreatePoint()

	eq, _ := point.Equal(point2)
	assert.False(t, eq)
}

func TestEd25519PointSet_NilParamShouldErr(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()

	err := point.Set(nil)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestEd25519PointSet_SavesCorrectValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()
	point2 := suite.CreatePoint()

	_ = point.Set(point2)
	eq, _ := point.Equal(point2)
	assert.True(t, eq)
}

func TestEd25519PointSet_CopiesValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()
	point2 := suite.CreatePoint()
	point3 := suite.CreatePoint()

	_ = point.Set(point2)
	_ = point2.Set(point3)
	eq, _ := point.Equal(point3)
	assert.False(t, eq)
}

func TestEd25519PointClone_ReturnsSameScalarValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()
	point2 := point.Clone()

	eq, _ := point.Equal(point2)
	assert.True(t, eq)
}

func TestEd25519PointClone_CopiesValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()
	point2 := point.Clone()
	point3 := suite.CreatePoint()
	_ = point.Set(point3)

	eq, _ := point2.Equal(point)
	assert.False(t, eq)
}

func TestEd25519PointMarshalUnmarshal(t *testing.T) {
	suite := ed25519.NewEd25519()
	point := suite.CreatePoint()

	bytes, _ := point.MarshalBinary()
	point2 := suite.CreatePoint()
	_ = point2.UnmarshalBinary(bytes)

	eq, _ := point.Equal(point2)
	assert.True(t, eq)
}
