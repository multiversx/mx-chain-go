package display

import "encoding/hex"

const ellipsisCharacter = "\u2026"

// ConvertHash generates a short-hand of provided bytes slice showing only the first 3 and the last 3 bytes as hex
// in total, the resulting string is maximum 13 characters long
func ConvertHash(hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	if len(hash) < 6 {
		return hex.EncodeToString(hash)
	}

	prefix := hex.EncodeToString(hash[:3])
	suffix := hex.EncodeToString(hash[len(hash)-3:])
	return prefix + ellipsisCharacter + suffix
}
