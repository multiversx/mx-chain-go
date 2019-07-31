package core

// Max returns the maximum number between two given
func Max(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
