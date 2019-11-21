package core

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

// PopUint32 removes the specified index from the provided slice and returns the resulting slice
//  If index is a negative number it will start from the tail. If the index absolute value
//  is out of the provided slice bouds it will return the full slice
func PopUint32(a []uint32, index int) []uint32 {
	l := len(a)
	if index < 0 {
		index = l + index
	}
	if index > l {
		return a
	}
	if index == l {
		return a[:index]
	}

	return append(a[:index], a[index+1:]...)
}
