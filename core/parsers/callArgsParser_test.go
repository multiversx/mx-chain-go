package parsers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCallArgsParser_ParseData(t *testing.T) {
	t.Parallel()

	parser := NewCallArgsParser()
	require.NotNil(t, parser)

	function, arguments, err := parser.ParseData("fooBar")
	require.Nil(t, err)
	require.Equal(t, "fooBar", function)
	require.Equal(t, [][]byte{}, arguments)

	function, arguments, err = parser.ParseData("fooBar@0A0A@0B0B")
	require.Nil(t, err)
	require.Equal(t, "fooBar", function)
	require.Equal(t, [][]byte{{10, 10}, {11, 11}}, arguments)
}

func TestCallArgsParser_ParseDataWhenErrorneousInput(t *testing.T) {
	t.Parallel()

	parser := NewCallArgsParser()
	require.NotNil(t, parser)

	function, arguments, err := parser.ParseData("")
	require.Equal(t, ErrTokenizeFailed, err)
	require.Equal(t, "", function)
	require.Nil(t, arguments)

	function, arguments, err = parser.ParseData("@a")
	require.Equal(t, ErrTokenizeFailed, err)
	require.Equal(t, "", function)
	require.Nil(t, arguments)

	function, arguments, err = parser.ParseData("foo@BADARG")
	require.Equal(t, ErrTokenizeFailed, err)
	require.Equal(t, "", function)
	require.Nil(t, arguments)
}
