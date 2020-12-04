package parsers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeployArgsParser_ParseData(t *testing.T) {
	t.Parallel()

	parser := NewDeployArgsParser()
	require.NotNil(t, parser)

	parsed, err := parser.ParseData("ABBA@0123@0000")
	require.Nil(t, err)
	require.NotNil(t, parsed)
	require.Equal(t, []byte{0xAB, 0xBA}, parsed.Code)
	require.Equal(t, []byte{0x01, 0x23}, parsed.VMType)
	require.False(t, parsed.CodeMetadata.Upgradeable)
	require.Equal(t, [][]byte{}, parsed.Arguments)

	parsed, err = parser.ParseData("ABBA@0123@0100@64@0A")
	require.Nil(t, err)
	require.NotNil(t, parsed)
	require.Equal(t, []byte{0xAB, 0xBA}, parsed.Code)
	require.Equal(t, []byte{0x01, 0x23}, parsed.VMType)
	require.True(t, parsed.CodeMetadata.Upgradeable)
	require.Equal(t, [][]byte{{100}, {0xA}}, parsed.Arguments)
}

func TestDeployArgsParser_ParseDataWhenErrorneousInput(t *testing.T) {
	t.Parallel()

	parser := NewDeployArgsParser()
	require.NotNil(t, parser)

	parsed, err := parser.ParseData("")
	require.Equal(t, ErrTokenizeFailed, err)
	require.Nil(t, parsed)

	parsed, err = parser.ParseData("@aaaa")
	require.Equal(t, ErrTokenizeFailed, err)
	require.Nil(t, parsed)

	parsed, err = parser.ParseData("ABBA@A")
	require.Equal(t, ErrInvalidDeployArguments, err)
	require.Nil(t, parsed)

	parsed, err = parser.ParseData("XYZY@A@A")
	require.Equal(t, ErrInvalidCode, err)
	require.Nil(t, parsed)

	parsed, err = parser.ParseData("ABBA@A@A")
	require.Equal(t, ErrInvalidVMType, err)
	require.Nil(t, parsed)

	parsed, err = parser.ParseData("ABBA@@A")
	require.Equal(t, ErrInvalidVMType, err)
	require.Nil(t, parsed)

	parsed, err = parser.ParseData("ABBA@ABBA@A")
	require.Equal(t, ErrInvalidCodeMetadata, err)
	require.Nil(t, parsed)

	parsed, err = parser.ParseData("ABBA@ABBA@ABBA@A")
	require.Equal(t, ErrTokenizeFailed, err)
	require.Nil(t, parsed)
}
