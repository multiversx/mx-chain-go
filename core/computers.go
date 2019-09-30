package core

import (
	"bytes"
)

// Max returns the maximum number between two given
func Max(a int32, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum number between two given
func Min(a int32, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// IsHashInList signals if the given hash exists in the given list of hashes
func IsHashInList(hash []byte, hashes [][]byte) bool {
	for i := 0; i < len(hashes); i++ {
		if bytes.Equal(hash, hashes[i]) {
			return true
		}
	}

	return false
}
