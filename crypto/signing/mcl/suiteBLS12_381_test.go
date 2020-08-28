package mcl

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSuiteBLS12(t *testing.T) {
	suite := NewSuiteBLS12()

	assert.NotNil(t, suite)
}

func TestSuiteBLS12_RandomStream(t *testing.T) {
	suite := NewSuiteBLS12()
	stream := suite.RandomStream()
	require.Nil(t, stream)
}

func TestSuiteBLS12_CreatePoint(t *testing.T) {
	suite := NewSuiteBLS12()

	point1 := suite.CreatePoint()
	point2 := suite.CreatePoint()

	assert.NotNil(t, point1)
	assert.NotNil(t, point2)
	assert.False(t, point1 == point2)
}

func TestSuiteBLS12_String(t *testing.T) {
	suite := NewSuiteBLS12()

	str := suite.String()
	assert.Equal(t, "BLS12-381 suite", str)
}

func TestSuiteBLS12_ScalarLen(t *testing.T) {
	suite := NewSuiteBLS12()

	length := suite.ScalarLen()
	assert.Equal(t, 32, length)
}

func TestSuiteBLS12_CreateScalar(t *testing.T) {
	suite := NewSuiteBLS12()

	scalar := suite.CreateScalar()
	assert.NotNil(t, scalar)
}

func TestSuiteBLS12_CreatePointForScalar(t *testing.T) {
	suite := NewSuiteBLS12()

	secretKey := &bls.SecretKey{}
	secretKey.SetByCSPRNG()
	secretKey.GetPublicKey()

	scalar := NewScalar()
	scalar.Scalar = bls.CastFromSecretKey(secretKey)

	point, err := suite.CreatePointForScalar(scalar)
	require.Nil(t, err)
	pG2, ok := point.GetUnderlyingObj().(*bls.G2)
	require.True(t, ok)
	require.NotNil(t, pG2)

	pubKey := secretKey.GetPublicKey()
	point2G2 := bls.CastFromPublicKey(pubKey)

	require.True(t, pG2.IsEqual(point2G2))
}

func TestSuiteBLS12_CreateKeyPair(t *testing.T) {
	suite := NewSuiteBLS12()

	scalar, point := suite.CreateKeyPair()
	mclScalar := scalar.GetUnderlyingObj().(*bls.Fr)

	secretKey := bls.CastToSecretKey(mclScalar)

	pG2, ok := point.GetUnderlyingObj().(*bls.G2)
	require.True(t, ok)
	require.NotNil(t, pG2)

	pubKey := secretKey.GetPublicKey()
	point2G2 := bls.CastFromPublicKey(pubKey)

	require.True(t, pG2.IsEqual(point2G2))
}

func TestSuiteBLS12_PointLen(t *testing.T) {
	suite := NewSuiteBLS12()

	pointLength := suite.PointLen()

	// G2 point length is 128 bytes
	assert.Equal(t, 96, pointLength)
}

func TestSuiteBLS12_CreateKey(t *testing.T) {
	suite := NewSuiteBLS12()
	private, public := suite.CreateKeyPair()
	assert.NotNil(t, private)
	assert.NotNil(t, public)
}

func TestSuiteBLS12_GetUnderlyingSuite(t *testing.T) {
	suite := NewSuiteBLS12()

	obj := suite.GetUnderlyingSuite()

	assert.NotNil(t, obj)
}

func TestSuiteBLS12_CheckPointValidOK(t *testing.T) {
	validPointHexStr := "368723d835fca6bc0c17a270e51b731f69f9fe482ed88e8c3d879f228291d48057aa12d0de8476b4a111e945399253" +
		"15d2d3fd1b85e29e465b8814b713cbf833115f4562e28dcf58e960751f0581578ca1819c8790aa5a5300c5c317b74dca01"

	suite := NewSuiteBLS12()

	validPointBytes, err := hex.DecodeString(validPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(validPointBytes)
	require.Nil(t, err)
}

func TestSuiteBLS12_CheckPointValidShortHexStringShouldErr(t *testing.T) {
	shortPointHexStr := "368723d835fca6bc0c17a270e51b731f69f9fe482ed88e8c3d879f228291d48057aa12d0de8476b4a111e945399253" +
		"15d2d3fd1b85e29e465b8814b713cbf833115f4562e28dcf58e960751f0581578ca1819c8790aa5a5300c5c317b74d"

	suite := NewSuiteBLS12()

	shortPointBytes, err := hex.DecodeString(shortPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(shortPointBytes)
	require.Equal(t, crypto.ErrInvalidParam, err)
}

func TestSuiteBLS12_CheckPointValidLongHexStrShouldErr(t *testing.T) {
	longPointHexStr := "368723d835fca6bc0c17a270e51b731f69f9fe482ed88e8c3d879f228291d48057aa12d0de8476b4a111e945399253" +
		"15d2d3fd1b85e29e465b8814b713cbf833115f4562e28dcf58e960751f0581578ca1819c8790aa5a5300c5c317b74d" +
		"15d2d3fd1b85e29e465b8814b713cbf833115f4562e28dcf58e960751f0581578ca1819c8790aa5a5300c5c317b74d"

	suite := NewSuiteBLS12()

	longPointBytes, err := hex.DecodeString(longPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(longPointBytes)
	require.Equal(t, crypto.ErrInvalidParam, err)
}

func TestSuiteBLS12_CheckPointValidInvalidPointHexStrShouldErr(t *testing.T) {
	invalidPointHexStr := "368723d835fca6bc0c17a270e51b731f69f9fe482ed88e8c3d879f228291d48057aa12d0de8476b4a111e945399253" +
		"15d2d3fd1b85e29e465b8814b713cbf833115f4562e28dcf58e960751f0581578ca1819c8790aa5a5300c5caaaaaaaaaaa"
	oneHexCharCorruptedPointHexStr := "368723d835fca6bc0c17a270e51b731f69f9fe482ed88e8c3d879f228291d48057aa12d0de8476b4a111e945399253" +
		"15d2d3fd1b85e29e465b8814b713cbf833115f4562e28dcf58e960751f0581578ca1819c8790aa5a5300c5c317b74dca0a"
	suite := NewSuiteBLS12()

	invalidPointBytes, err := hex.DecodeString(invalidPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(invalidPointBytes)
	require.NotNil(t, err)

	oneHexCharCorruptedPointBytes, err := hex.DecodeString(oneHexCharCorruptedPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(oneHexCharCorruptedPointBytes)
	require.NotNil(t, err)
}

func TestSuiteBLS12_CheckPointValidZeroHexStrShouldErr(t *testing.T) {
	zeroPointHexStr := "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

	suite := NewSuiteBLS12()

	zeroPointBytes, err := hex.DecodeString(zeroPointHexStr)
	require.Nil(t, err)
	err = suite.CheckPointValid(zeroPointBytes)
	require.Equal(t, crypto.ErrInvalidPoint, err)
}

func TestSuiteBLS12_IsInterfaceNil(t *testing.T) {
	t.Parallel()
	var suite *SuiteBLS12

	require.True(t, check.IfNil(suite))
	suite = NewSuiteBLS12()
	require.False(t, check.IfNil(suite))
}
