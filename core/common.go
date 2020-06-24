package core

import (
	"crypto/rand"
	"time"
)

// TimeToGracefullyCloseNode is used as a delay before forcefully closing the node after the gracefully close was called
const TimeToGracefullyCloseNode = time.Second * 10

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
