package factory

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/legacyKeccak256"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
)

// NewHasher will return a new instance of hasher based on the value stored in config
func NewHasher(name string) (hashing.Hasher, error) {
	switch name {
	case "sha256":
		return sha256.Sha256{}, nil
	case "legacyKeccak256":
		return &legacyKeccak256.LegacyKeccak256{}, nil
	case "blake2b":
		return &blake2b.Blake2b{}, nil
	}

	return nil, ErrNoHasherInConfig
}
