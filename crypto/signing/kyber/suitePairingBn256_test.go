package kyber_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/stretchr/testify/assert"
)

func TestNewSuitePairingBn256(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	assert.NotNil(t, suite)
}

func TestSuitePairingBn256_RandomStream(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()
	stream := suite.RandomStream()

	assert.NotNil(t, stream)
}

func TestSuitePairingBn256_CreatePoint(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	point1 := suite.CreatePoint()
	point2 := suite.CreatePoint()

	assert.NotNil(t, point1)
	assert.NotNil(t, point2)
	assert.False(t, point1 == point2)
}

func TestSuitePairingBn256_String(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	str := suite.String()
	assert.Equal(t, "bn256.adapter", str)
}

func TestSuitePairingBn256_ScalarLen(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	length := suite.ScalarLen()
	assert.Equal(t, 32, length)
}

func TestSuitePairingBn256_CreateScalar(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	scalar := suite.CreateScalar()
	assert.NotNil(t, scalar)
}

func TestSuitePairingBn256_PointLen(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	pointLength := suite.PointLen()

	// G2 point length is 128 bytes
	assert.Equal(t, 128, pointLength)
}

func TestSuitePairingBn256_CreateKey(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	stream := suite.RandomStream()
	private, public := suite.CreateKeyPair(stream)

	assert.NotNil(t, private)
	assert.NotNil(t, public)
}

func TestSuitePairingBn256_GetUnderlyingSuite(t *testing.T) {
	suite := kyber.NewSuitePairingBn256()

	obj := suite.GetUnderlyingSuite()

	assert.NotNil(t, obj)
}
