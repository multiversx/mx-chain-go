package sliceUtil

// TrimSliceSliceByte creates a copy of the provided slice without the excess capacity
func TrimSliceSliceByte(in [][]byte) [][]byte {
	if len(in) == 0 {
		return [][]byte{}
	}
	ret := make([][]byte, len(in))
	copy(ret, in)
	return ret
}
