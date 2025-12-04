package helpers

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/multiversx/mx-chain-core-go/data/block"
)

// ComputeRandomnessForCleanup returns the randomness for bad transactions removal
// TODO Maybe find a simpler solution for randomness calculation
func ComputeRandomnessForCleanup(body *block.Body) uint64 {
	randomness := uint64(0)

	if len(body.MiniBlocks) != 0 && len(body.MiniBlocks[0].TxHashes) != 0 {
		firstHash := body.MiniBlocks[0].TxHashes[len(body.MiniBlocks[0].TxHashes)-1]
		hashSum := sha256.Sum256(firstHash)

		randomness = binary.LittleEndian.Uint64(hashSum[:8]) % uint64(len(body.MiniBlocks[0].TxHashes))
	}

	return randomness
}
