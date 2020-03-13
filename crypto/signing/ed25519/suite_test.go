package ed25519_test

import (
	goEd25519 "crypto/ed25519"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestNewEd25519(t *testing.T) {
	suite := ed25519.NewEd25519()
	assert.False(t, check.IfNil(suite))
}

func TestNewEd25519CreateKeyPair(t *testing.T) {
	suite := ed25519.NewEd25519()
	privateKey, publicKey := suite.CreateKeyPair(nil)
	assert.NotNil(t, privateKey)
	assert.NotNil(t, publicKey)
}

func TestNewEd25519CreateKeyPair_GeneratesDifferentKeys(t *testing.T) {
	suite := ed25519.NewEd25519()
	privateKey, publicKey := suite.CreateKeyPair(nil)
	privateKey2, publicKey2 := suite.CreateKeyPair(nil)

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
