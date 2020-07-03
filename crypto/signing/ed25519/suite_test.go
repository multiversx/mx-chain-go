package ed25519_test

import (
	goEd25519 "crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEd25519(t *testing.T) {
	suite := ed25519.NewEd25519()
	assert.False(t, check.IfNil(suite))
}

func TestNewEd25519CreateKeyPair(t *testing.T) {
	suite := ed25519.NewEd25519()
	privateKey, publicKey := suite.CreateKeyPair()
	assert.NotNil(t, privateKey)
	assert.NotNil(t, publicKey)
}

func TestNewEd25519CreateKeyPair_GeneratesDifferentKeys(t *testing.T) {
	suite := ed25519.NewEd25519()
	privateKey, publicKey := suite.CreateKeyPair()
	privateKey2, publicKey2 := suite.CreateKeyPair()

	assert.NotEqual(t, privateKey, privateKey2)
	assert.NotEqual(t, publicKey, publicKey2)
}

func TestNewEd25519CreatePoint(t *testing.T) {
	suite := ed25519.NewEd25519()
	publicKey := suite.CreatePoint()
	assert.NotNil(t, publicKey)
}

func TestNewEd25519CreateScalar(t *testing.T) {
	suite := ed25519.NewEd25519()
	privateKey := suite.CreateScalar()
	assert.NotNil(t, privateKey)
}

func TestNewEd25519String(t *testing.T) {
	suite := ed25519.NewEd25519()
	assert.Equal(t, ed25519.ED25519, suite.String())
}

func TestNewEd25519ScalarLen(t *testing.T) {
	suite := ed25519.NewEd25519()
	assert.Equal(t, goEd25519.PrivateKeySize, suite.ScalarLen())
}

func TestNewEd25519PointLen(t *testing.T) {
	suite := ed25519.NewEd25519()
	assert.Equal(t, goEd25519.PublicKeySize, suite.PointLen())
}

func TestSuiteEd25519_CheckPointValid(t *testing.T) {
	validPointHexStr := "246008bbf5ebb46892c4b079c4ba5d76ee2d5f648ab8005ff082029c8e8daa18"
	shortPointHexStr := "246008bbf5ebb46892c4b079c4ba5d76ee2d5f648ab8005ff082029c8e8daa"
	longPointHexStr := "246008bbf5ebb46892c4b079c4ba5d76ee2d5f648ab8005ff082029c8e8daa1818"

	suite := ed25519.NewEd25519()

	validPointBytes, err := hex.DecodeString(validPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(validPointBytes)
	require.Nil(t, err)

	shortPointBytes, err := hex.DecodeString(shortPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(shortPointBytes)
	require.Equal(t, crypto.ErrInvalidParam, err)

	longPointBytes, err := hex.DecodeString(longPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(longPointBytes)
	require.Equal(t, crypto.ErrInvalidParam, err)
}
