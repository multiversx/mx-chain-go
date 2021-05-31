package core

import (
	"crypto/rand"
	"os"
)

// EmptyChannel empties the given channel
func EmptyChannel(ch chan bool) int {
	readsCnt := 0
	for {
		select {
		case <-ch:
			readsCnt++
		default:
			return readsCnt
		}
	}
}

// UniqueIdentifier returns a unique string identifier of 32 bytes
func UniqueIdentifier() string {
	buff := make([]byte, 32)
	_, _ = rand.Read(buff)
	return string(buff)
}

// FileExists returns true if the file at the given path exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetPBFTThreshold returns the pBFT threshold for a given consensus size
func GetPBFTThreshold(consensusSize int) int {
	return consensusSize*2/3 + 1
}

// GetPBFTFallbackThreshold returns the pBFT fallback threshold for a given consensus size
func GetPBFTFallbackThreshold(consensusSize int) int {
	return consensusSize*1/2 + 1
}
