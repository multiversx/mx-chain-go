package miniblocks

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareBufferMiniblocks(t *testing.T) {
	buff := &bytes.Buffer{}

	meta := []byte("test1")
	serializedData := []byte("test2")

	putInBufferMiniblockData(buff, meta, serializedData)

	expectedBuff := &bytes.Buffer{}
	serializedData = append(serializedData, "\n"...)
	expectedBuff.Grow(len(meta) + len(serializedData))
	_, _ = expectedBuff.Write(meta)
	_, _ = expectedBuff.Write(serializedData)

	require.Equal(t, expectedBuff, buff)
}
