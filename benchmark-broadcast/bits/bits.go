package bits

import "fmt"

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
