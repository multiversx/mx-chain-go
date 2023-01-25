package trie

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
)

func TestNewMerkleProofVerifier_NilMarshalizer(t *testing.T) {
	t.Parallel()

	mpv, err := NewMerkleProofVerifier(nil, sha256.NewSha256())
	assert.Nil(t, mpv)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestNewMerkleProofVerifier_NilHasher(t *testing.T) {
	t.Parallel()

	mpv, err := NewMerkleProofVerifier(&marshal.GogoProtoMarshalizer{}, nil)
	assert.Nil(t, mpv)
	assert.Equal(t, ErrNilHasher, err)
}

func TestNewMerkleProofVerifier(t *testing.T) {
	t.Parallel()

	mpv, err := NewMerkleProofVerifier(&marshal.GogoProtoMarshalizer{}, sha256.NewSha256())
	assert.Nil(t, err)
	assert.NotNil(t, mpv)
}

func TestMerkleProofVerifier_VerifyProof(t *testing.T) {
	t.Parallel()

	mpv, _ := NewMerkleProofVerifier(&marshal.GogoProtoMarshalizer{}, sha256.NewSha256())

	rootHash := []byte{188, 46, 84, 157, 152, 195, 31, 254, 110, 148, 25, 185, 51, 208, 59, 55, 232, 79, 116, 196, 38, 1, 65, 35, 2, 121, 157, 39, 118, 81, 166, 216}
	address := []byte{191, 66, 33, 55, 71, 105, 126, 157, 236, 66, 17, 239, 80, 186, 96, 97, 181, 71, 41, 181, 59, 160, 196, 153, 73, 72, 202, 180, 120, 175, 136, 84}
	p, _ := hex.DecodeString("0a41040508080f0a0807040b0a0c080409040909040c000a0b03050b09020704050b010600060a0b00050f0e010102040c0e0d090e07090607040703010202040f0b10124c1202000022206182d14320be95434f5508acad9478d3b6cf837bfce7ebfe47c2e860d1b98ca72a20bf42213747697e9dec4211ef50ba6061b54729b53ba0c4994948cab478af88543202000001")
	proof := [][]byte{p}

	ok, err := mpv.VerifyProof(rootHash, address, proof)
	assert.Nil(t, err)
	assert.True(t, ok)
}
