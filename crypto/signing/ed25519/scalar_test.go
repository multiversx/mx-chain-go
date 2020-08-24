package ed25519_test

import (
	goEd25519 "crypto/ed25519"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestEd25519ScalarEqual_NilParamShouldErr(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()

	_, err := scalar.Equal(nil)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestEd25519ScalarEqual_InvalidParamShouldErr(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()

	scalar2 := &mock.ScalarMock{}
	_, err := scalar.Equal(scalar2)
	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestEd25519ScalarEqual_ReturnsTrueForTheSameKey(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()

	eq, _ := scalar.Equal(scalar)
	assert.True(t, eq)
}

func TestEd25519ScalarEqual_ReturnsFalseForDifferentKeys(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()
	scalar2 := suite.CreateScalar()

	eq, _ := scalar.Equal(scalar2)
	assert.False(t, eq)
}

func TestEd25519ScalarSet_NilParamShouldErr(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()

	err := scalar.Set(nil)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestEd25519ScalarSet_SavesCorrectValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()
	scalar2 := suite.CreateScalar()

	_ = scalar.Set(scalar2)
	eq, _ := scalar.Equal(scalar2)
	assert.True(t, eq)
}

func TestEd25519ScalarSet_CopiesValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()
	scalar2 := suite.CreateScalar()
	scalar3 := suite.CreateScalar()

	_ = scalar.Set(scalar2)
	_ = scalar2.Set(scalar3)
	eq, _ := scalar.Equal(scalar3)
	assert.False(t, eq)
}

func TestEd25519ScalarClone_ReturnsSameScalarValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()
	scalar2 := scalar.Clone()

	eq, _ := scalar.Equal(scalar2)
	assert.True(t, eq)
}

func TestEd25519ScalarClone_CopiesValue(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()
	scalar2 := scalar.Clone()
	scalar3 := suite.CreateScalar()
	_ = scalar.Set(scalar3)

	eq, _ := scalar2.Equal(scalar)
	assert.False(t, eq)
}

func TestEd25519MarshalBinary_WrongKeyType(t *testing.T) {
	scalar := ed25519.NewScalar([]byte("wrong key"))
	_, err := scalar.MarshalBinary()

	assert.Equal(t, crypto.ErrWrongPrivateKeySize, err)
}

func TestEd25519ScalarMarshalUnmarshal(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()

	bytes, _ := scalar.MarshalBinary()
	scalar2 := suite.CreateScalar()
	_ = scalar2.UnmarshalBinary(bytes)

	eq, _ := scalar.Equal(scalar2)
	assert.True(t, eq)
}

func TestEd25519ScalarUnmarshal_WorksWithSeed(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()
	scalar2 := suite.CreateScalar()

	bytes, _ := scalar.MarshalBinary()
	_ = scalar2.UnmarshalBinary(bytes[:32])

	eq, _ := scalar.Equal(scalar2)
	assert.True(t, eq)
}

func TestEd25519ScalarUnmarshal_ErrorOnWrongSize(t *testing.T) {
	suite := ed25519.NewEd25519()
	scalar := suite.CreateScalar()

	err := scalar.UnmarshalBinary([]byte("wrong size"))

	assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
}

func TestEd255192IsKeyValid_ErrOnWrongSize(t *testing.T) {
	privateKey := []byte("wrong size")
	err := ed25519.IsKeyValid(privateKey)

	assert.Equal(t, crypto.ErrWrongPrivateKeySize, err)
}

func TestEd255192IsKeyValid_ErrOnWrongStructure(t *testing.T) {
	suite := ed25519.NewEd25519()
	privateKey, _ := suite.CreateKeyPair()
	privateKeyBytes, _ := (privateKey.GetUnderlyingObj()).(goEd25519.PrivateKey)
	privateKeyBytes[len(privateKeyBytes)-1]++

	err := ed25519.IsKeyValid(privateKeyBytes)
	assert.Equal(t, crypto.ErrWrongPrivateKeyStructure, err)
}

func TestEd255192IsKeyValid_CorrectKey(t *testing.T) {
	suite := ed25519.NewEd25519()
	privateKey, _ := suite.CreateKeyPair()
	privateKeyBytes, _ := (privateKey.GetUnderlyingObj()).(goEd25519.PrivateKey)

	err := ed25519.IsKeyValid(privateKeyBytes)
	assert.Nil(t, err)
}

func TestEd25519GetUnderlyingObj_InvalidKey(t *testing.T) {
	scalar := ed25519.NewScalar([]byte("wrong size"))
	privateKey := scalar.GetUnderlyingObj()

	assert.Nil(t, privateKey)
}
