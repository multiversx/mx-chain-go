package mock

import (
	"encoding/hex"
)

type Streamer struct {
	key []byte
}

func NewStreamer() *Streamer {
	key, _ := hex.DecodeString("aa")
	return &Streamer{key: key}
}

func (stream *Streamer) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("dst length < src length")
	}

	for i := 0; i < len(src); i++ {
		dst[i] = src[i] ^ stream.key[0]
	}
}
