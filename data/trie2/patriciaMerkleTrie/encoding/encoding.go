package encoding

import (
	"bytes"
)

// KeyBytesToHex turns key bytes into hex nibbles.
func KeyBytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}

// PrefixLen returns the length of the common prefix of a and b.
func PrefixLen(a, b []byte) int {
	var i, length = 0, len(a)
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

// HasPrefix tests whether the byte slice s begins with prefix.
func HasPrefix(s, prefix []byte) bool {
	return len(s) >= len(prefix) && bytes.Equal(s[0:len(prefix)], prefix)
}

// HasTerm returns whether a hex key has the terminator flag.
func HasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == 16
}

// HexToKeyBytes turns hex nibbles into key bytes.
// This can only be used for keys of even length.
func HexToKeyBytes(hex []byte) []byte {
	if HasTerm(hex) {
		hex = hex[:len(hex)-1]
	}
	if len(hex)&1 != 0 {
		panic("can't convert hex key of odd length")
	}
	key := make([]byte, len(hex)/2)

	for bi, ni := 0, 0; ni < len(hex); bi, ni = bi+1, ni+2 {
		key[bi] = hex[ni]<<4 | hex[ni+1]
	}

	return key
}
