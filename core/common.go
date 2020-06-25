package core

import (
	"crypto/rand"
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
