package core

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// ConvertBytes converts the input bytes in a readable string using multipliers (k, M, G)
func ConvertBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f kiB", float64(bytes)/1024.0)
	}
	if bytes < 1024*1024*1025 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024.0/1024.0)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/1024.0/1024.0/1024.0)
}

// ToB64 encodes the given buff to base64
func ToB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

// ToHex encodes the given buff to hex
func ToHex(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return "0x" + hex.EncodeToString(buff)
}
