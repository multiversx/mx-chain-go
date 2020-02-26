package rand

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

type seedRandReader struct {
	seedNumber int64
}

// NewSeedRandReader will return a new instance of a seed-based reader
// This is mostly used to generate predictable seeder addresses so other peers can connect to
func NewSeedRandReader(seed []byte) (*seedRandReader, error) {
	if len(seed) == 0 {
		return nil, p2p.ErrEmptySeed
	}

	seedHash := sha256.Sum256(seed)
	seedNumber := binary.BigEndian.Uint64(seedHash[:])

	return &seedRandReader{
		seedNumber: int64(seedNumber),
	}, nil
}

// Read will read upto len(p) bytes. It will rotate the existing byte buffer (seed) until it will fill up the provided
// p buffer
func (srr *seedRandReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, p2p.ErrEmptyBuffer
	}

	randomizer := rand.New(rand.NewSource(srr.seedNumber))

	return randomizer.Read(p)
}
