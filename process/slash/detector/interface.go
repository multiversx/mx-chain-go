package detector

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// RoundValidatorHeadersCache defines what a <round, <pubKey, headers>> (<key1, <key2, val>>) cache should do.
type RoundValidatorHeadersCache interface {
	// Add should add in cache a header info data for a public key, in a given round.
	// If the public key already has the header cached in the given round OR the round is
	// irrelevant(obsolete) to be cached, an error should be returned.
	// If the cache is full, it should have an eviction mechanism to always remove
	// the oldest round entry
	Add(round uint64, pubKey []byte, headerInfo data.HeaderInfoHandler) error
	// GetHeaders returns all cached headers for a public key, in a given round
	GetHeaders(round uint64, pubKey []byte) []data.HeaderInfoHandler
	// GetPubKeys returns all cached public keys in a given round
	GetPubKeys(round uint64) [][]byte
	// IsInterfaceNil checks if the interface is nil
	IsInterfaceNil() bool
}

// RoundHashCache defines what a <round, hash>(<key, val>) cache should do
type RoundHashCache interface {
	// Add should add a hash in cache for a given round.
	// If the hash is already cached in the given round OR the
	// round is irrelevant(obsolete) to be cached, an error should be returned.
	// If the cache is full, it should have an eviction mechanism
	// to always remove the oldest round entry
	Add(round uint64, hash []byte) error
	// Remove should remove a hash, from the list of hashes, in a given round
	Remove(round uint64, hash []byte)
	// IsInterfaceNil checks if the interface is nil
	IsInterfaceNil() bool
}
