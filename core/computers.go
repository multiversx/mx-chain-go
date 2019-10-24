package core

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
)

// MaxInt32 returns the maximum of two given numbers
func MaxInt32(a int32, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// MinInt32 returns the minimum of two given numbers
func MinInt32(a int32, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// MaxUint32 returns the maximum of two given numbers
func MaxUint32(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// MinUint32 returns the minimum of two given numbers
func MinUint32(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// MaxUint64 returns the maximum of two given numbers
func MaxUint64(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// MinUint64 returns the minimum of two given numbers
func MinUint64(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func IsSmartContractAddress(rcvAddress []byte) bool {
	isEmptyAddress := bytes.Equal(rcvAddress, make([]byte, len(rcvAddress)))
	if isEmptyAddress {
		return true
	}

	isSCAddress := bytes.Equal(rcvAddress[:(hooks.NumInitCharactersForScAddress-hooks.VMTypeLen)],
		make([]byte, hooks.NumInitCharactersForScAddress-hooks.VMTypeLen))
	if isSCAddress {
		return true
	}

	return false
}
