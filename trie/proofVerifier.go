package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type merkleProofVerifier struct {
	trie *patriciaMerkleTrie
}

// NewMerkleProofVerifier creates a new instance of merkleProofVerifier
func NewMerkleProofVerifier(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*merkleProofVerifier, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &merkleProofVerifier{
		trie: &patriciaMerkleTrie{
			marshalizer: marshalizer,
			hasher:      hasher,
		},
	}, nil
}

// VerifyProof verifies the given Merkle proof
func (mpv *merkleProofVerifier) VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error) {
	return mpv.trie.VerifyProof(rootHash, key, proof)
}
