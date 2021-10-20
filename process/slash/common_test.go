package slash_test

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/stretchr/testify/require"
)

func TestIsIndexSetInBitmap(t *testing.T) {
	byte1Map, _ := strconv.ParseInt("11001101", 2, 9)
	byte2Map, _ := strconv.ParseInt("00000101", 2, 9)
	bitmap := []byte{byte(byte1Map), byte(byte2Map)}

	//Byte 1
	require.True(t, slash.IsIndexSetInBitmap(0, bitmap))
	require.False(t, slash.IsIndexSetInBitmap(1, bitmap))
	require.True(t, slash.IsIndexSetInBitmap(2, bitmap))
	require.True(t, slash.IsIndexSetInBitmap(3, bitmap))
	require.False(t, slash.IsIndexSetInBitmap(4, bitmap))
	require.False(t, slash.IsIndexSetInBitmap(5, bitmap))
	require.True(t, slash.IsIndexSetInBitmap(6, bitmap))
	require.True(t, slash.IsIndexSetInBitmap(7, bitmap))
	// Byte 2
	require.True(t, slash.IsIndexSetInBitmap(8, bitmap))
	require.False(t, slash.IsIndexSetInBitmap(9, bitmap))
	require.True(t, slash.IsIndexSetInBitmap(10, bitmap))

	for i := uint32(11); i <= 100; i++ {
		require.False(t, slash.IsIndexSetInBitmap(i, bitmap))
	}
}
