package bits

import "fmt"

// XOR will calculate and return the xor operation result for the two given parameters
func XOR(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("length of byte slices is not equivalent: %d != %d", len(a), len(b))
	}

	buf := make([]byte, len(a))

	for i := range a {
		buf[i] = a[i] ^ b[i]
	}

	return buf, nil
}

// CountSet returns the number of bits of 1 in a byte slice
func CountSet(nr []byte) int {
	count := 0
	for _, byteNo := range nr {
		for i := 0; i < 8; i++ {
			if byteNo&1 == 1 {
				count++
			}
			byteNo >>= 1
		}
	}
	return count
}
