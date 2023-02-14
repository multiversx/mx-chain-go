package holders

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type blockInfo struct {
	nonce    uint64
	hash     []byte
	rootHash []byte
}

// NewBlockInfo returns a new instance of a blockInfo struct
func NewBlockInfo(hash []byte, nonce uint64, rootHash []byte) *blockInfo {
	return &blockInfo{
		nonce:    nonce,
		hash:     hash,
		rootHash: rootHash,
	}
}

// GetNonce returns the block's nonce
func (bi *blockInfo) GetNonce() uint64 {
	return bi.nonce
}

// GetHash returns the block's hash
func (bi *blockInfo) GetHash() []byte {
	return bi.hash
}

// GetRootHash retuns the block's root hash
func (bi *blockInfo) GetRootHash() []byte {
	return bi.rootHash
}

// Equal checks if all inner fields are equal
func (bi *blockInfo) Equal(blockInfo common.BlockInfo) bool {
	if check.IfNil(blockInfo) {
		return false
	}

	areEqual := bi.nonce == blockInfo.GetNonce() &&
		bytes.Equal(bi.hash, blockInfo.GetHash()) &&
		bytes.Equal(bi.rootHash, blockInfo.GetRootHash())

	return areEqual
}

// IsInterfaceNil returns true if there is no value under the interface
func (bi *blockInfo) IsInterfaceNil() bool {
	return bi == nil
}
