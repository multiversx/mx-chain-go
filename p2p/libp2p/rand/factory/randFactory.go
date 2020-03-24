package factory

import (
	cryptoRand "crypto/rand"
	"io"

	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/rand"
)

// NewRandFactory will create a reader based on the provided seed string
func NewRandFactory(seed string) (io.Reader, error) {
	if len(seed) == 0 {
		return cryptoRand.Reader, nil
	}

	return rand.NewSeedRandReader([]byte(seed))
}
