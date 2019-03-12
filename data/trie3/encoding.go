package trie3

func keyBytesToHex(str []byte) []byte {
	length := len(str)*2 + 1
	var nibbles = make([]byte, length)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[length-1] = 16
	return nibbles
}

func prefixLen(a, b []byte) int {
	i := 0
	length := len(a)
	if len(b) < length {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

// HasTerm returns whether a hex key has the terminator flag.
func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == 16
}

// HexToKeyBytes turns hex nibbles into key bytes.
// This can only be used for keys of even length.
func hexToKeyBytes(hex []byte) []byte {
	if hasTerm(hex) {
		hex = hex[:len(hex)-1]
	}
	if len(hex)&1 != 0 {
		return nil
	}
	key := make([]byte, len(hex)/2)

	for bi, ni := 0, 0; ni < len(hex); bi, ni = bi+1, ni+2 {
		key[bi] = hex[ni]<<4 | hex[ni+1]
	}

	return key
}
