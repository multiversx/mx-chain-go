package vmcommon

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodeMetadata_FromBytes(t *testing.T) {
	require.True(t, CodeMetadataFromBytes([]byte{1, 0}).Upgradeable)
	require.False(t, CodeMetadataFromBytes([]byte{1, 0}).Readable)
	require.True(t, CodeMetadataFromBytes([]byte{0, 2}).Payable)
	require.True(t, CodeMetadataFromBytes([]byte{4, 0}).Readable)
	require.False(t, CodeMetadataFromBytes([]byte{4, 0}).Upgradeable)
	require.False(t, CodeMetadataFromBytes([]byte{0, 0}).Upgradeable)
	require.False(t, CodeMetadataFromBytes([]byte{0, 0}).Payable)
	require.False(t, CodeMetadataFromBytes([]byte{0, 0}).Readable)
}

func TestCodeMetadata_ToBytes(t *testing.T) {
	require.Equal(t, byte(0), (&CodeMetadata{}).ToBytes()[0])
	require.Equal(t, byte(0), (&CodeMetadata{}).ToBytes()[1])
	require.Equal(t, byte(1), (&CodeMetadata{Upgradeable: true}).ToBytes()[0])
	require.Equal(t, byte(2), (&CodeMetadata{Payable: true}).ToBytes()[1])
	require.Equal(t, byte(4), (&CodeMetadata{Readable: true}).ToBytes()[0])
}
