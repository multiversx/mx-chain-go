package core

// MaxInt32 returns the maximum number between two given
func MaxInt32(a int32, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// MinInt32 returns the minimum number between two given
func MinInt32(a int32, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// MaxUint32 returns the maximum number between two given
func MaxUint32(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// MinUint32 returns the minimum number between two given
func MinUint32(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
