package smartContract

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodeMetadata_FromBytes(t *testing.T) {
	require.True(t, CodeMetadataFromBytes([]byte{1}).Upgradeable)
	require.False(t, CodeMetadataFromBytes([]byte{0}).Upgradeable)
}

func TestCodeMetadata_ToBytes(t *testing.T) {
	require.Equal(t, byte(0), (&CodeMetadata{}).ToBytes()[0])
	require.Equal(t, byte(1), (&CodeMetadata{Upgradeable: true}).ToBytes()[0])
}
