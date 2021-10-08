package detector

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// RoundDetectorCache defines what a round-based(per validator data) cache should do.
type RoundDetectorCache interface {
	// Add should add in cache an intercepted data for a public key, in a given round.
	// If the cache is full, it should have an eviction mechanism to always remove
	// the oldest round entry
	Add(round uint64, pubKey []byte, data process.InterceptedData)
	// Contains checks if a public key, in a given round has any data cached.
	// It should do so by simply checking the hashes
	Contains(round uint64, pubKey []byte, data process.InterceptedData) bool
	// GetData returns all cached data for a public key, in a given round
	GetData(round uint64, pubKey []byte) []process.InterceptedData
	// GetPubKeys returns all cached public keys for a given round
	GetPubKeys(round uint64) [][]byte
	// IsInterfaceNil checks if the interface is nil
	IsInterfaceNil() bool
}

// HeadersCache defines what a header-hash-based cache should do
type HeadersCache interface {
	// Add should add in cache a header, along with its hash.
	// If the cache is full, it should have an eviction mechanism
	// to always remove the oldest round entry
	Add(round uint64, hash []byte, header data.HeaderHandler)
	// Contains checks if the hash in the given round is cached
	Contains(round uint64, hash []byte) bool
	// IsInterfaceNil checks if the interface is nil
	IsInterfaceNil() bool
}
