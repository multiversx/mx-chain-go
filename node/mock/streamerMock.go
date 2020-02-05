package mock

import (
	"encoding/hex"
)

// Streamer -
type Streamer struct {
	key []byte
}

// NewStreamer -
func NewStreamer() *Streamer {
	key, _ := hex.DecodeString("aa")
	return &Streamer{key: key}
}

// XORKeyStream -
func (stream *Streamer) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("dst length < src length")
	}

	for i := 0; i < len(src); i++ {
		dst[i] = src[i] ^ stream.key[0]
	}
}
