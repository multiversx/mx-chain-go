package builtInFunctions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestESDTGlobalMetaData_ToBytesWhenPaused(t *testing.T) {
	t.Parallel()

	esdtMetaData := &ESDTGlobalMetadata{
		Paused: true,
	}

	expected := make([]byte, lengthOfESDTMetadata)
	expected[0] = 1
	actual := esdtMetaData.ToBytes()
	require.Equal(t, expected, actual)
}

func TestESDTGlobalMetaData_ToBytesWhenNotPaused(t *testing.T) {
	t.Parallel()

	esdtMetaData := &ESDTGlobalMetadata{
		Paused: false,
	}

	expected := make([]byte, lengthOfESDTMetadata)
	expected[0] = 0
	actual := esdtMetaData.ToBytes()
	require.Equal(t, expected, actual)
}

func TestESDTGlobalMetadataFromBytes_InvalidLength(t *testing.T) {
	t.Parallel()

	emptyEsdtGlobalMetaData := ESDTGlobalMetadata{}

	invalidLengthByteSlice := make([]byte, lengthOfESDTMetadata+1)

	result := ESDTGlobalMetadataFromBytes(invalidLengthByteSlice)
	require.Equal(t, emptyEsdtGlobalMetaData, result)
}

func TestESDTGlobalMetadataFromBytes_ShouldSetPausedToTrue(t *testing.T) {
	t.Parallel()

	input := make([]byte, lengthOfESDTMetadata)
	input[0] = 1

	result := ESDTGlobalMetadataFromBytes(input)
	require.True(t, result.Paused)
}

func TestESDTGlobalMetadataFromBytes_ShouldSetPausedToFalse(t *testing.T) {
	t.Parallel()

	input := make([]byte, lengthOfESDTMetadata)
	input[0] = 0

	result := ESDTGlobalMetadataFromBytes(input)
	require.False(t, result.Paused)
}

func TestESDTUserMetaData_ToBytesWhenFrozen(t *testing.T) {
	t.Parallel()

	esdtMetaData := &ESDTUserMetadata{
		Frozen: true,
	}

	expected := make([]byte, lengthOfESDTMetadata)
	expected[0] = 1
	actual := esdtMetaData.ToBytes()
	require.Equal(t, expected, actual)
}

func TestESDTUserMetaData_ToBytesWhenNotFrozen(t *testing.T) {
	t.Parallel()

	esdtMetaData := &ESDTUserMetadata{
		Frozen: false,
	}

	expected := make([]byte, lengthOfESDTMetadata)
	expected[0] = 0
	actual := esdtMetaData.ToBytes()
	require.Equal(t, expected, actual)
}

func TestESDTUserMetadataFromBytes_InvalidLength(t *testing.T) {
	t.Parallel()

	emptyEsdtUserMetaData := ESDTUserMetadata{}

	invalidLengthByteSlice := make([]byte, lengthOfESDTMetadata+1)

	result := ESDTUserMetadataFromBytes(invalidLengthByteSlice)
	require.Equal(t, emptyEsdtUserMetaData, result)
}

func TestESDTUserMetadataFromBytes_ShouldSetFrozenToTrue(t *testing.T) {
	t.Parallel()

	input := make([]byte, lengthOfESDTMetadata)
	input[0] = 1

	result := ESDTUserMetadataFromBytes(input)
	require.True(t, result.Frozen)
}

func TestESDTUserMetadataFromBytes_ShouldSetFrozenToFalse(t *testing.T) {
	t.Parallel()

	input := make([]byte, lengthOfESDTMetadata)
	input[0] = 0

	result := ESDTUserMetadataFromBytes(input)
	require.False(t, result.Frozen)
}
