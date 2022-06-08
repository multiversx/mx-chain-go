package state

type accountBlockInfo struct {
	nonce    uint64
	hash     []byte
	rootHash []byte
}

// GetNonce returns the block nonce
func (info *accountBlockInfo) GetNonce() uint64 {
	return info.nonce
}

// GetHash returns the block hash
func (info *accountBlockInfo) GetHash() []byte {
	return info.hash
}

// GetRootHash returns the block rootHash
func (info *accountBlockInfo) GetRootHash() []byte {
	return info.rootHash
}

// IsInterfaceNil returns true if there is no value under the interface
func (info *accountBlockInfo) IsInterfaceNil() bool {
	return info == nil
}
