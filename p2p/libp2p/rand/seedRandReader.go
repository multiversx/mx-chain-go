package rand

import "github.com/ElrondNetwork/elrond-go/p2p"

type seedRandReader struct {
	seed []byte
}

// NewSeedRandReader will return a new instance of a seed-based reader
// This is mostly used to generate predictable seeder addresses so other peers can connect to
func NewSeedRandReader(seed []byte) (*seedRandReader, error) {
	if len(seed) == 0 {
		return nil, p2p.ErrEmptySeed
	}
	return &seedRandReader{
		seed: seed,
	}, nil
}

// Read will read upto len(p) bytes. It will rotate the existing byte buffer (seed) until it will fill up the provided
// p buffer
func (srr *seedRandReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, p2p.ErrEmptyBuffer
	}

	for i := 0; i < len(p); i++ {
		idx := i % len(srr.seed)
		p[i] = srr.seed[idx]
	}

	return len(p), nil
}
