package holders

type blockchainInfo struct {
	latestExecutedBlockHash []byte
	latestCommitedBlockHash []byte
	currentNonce            uint64
}

// NewBlockchainInfo creates a new instance of blockchainInfo
func NewBlockchainInfo(
	latestExecutedBlockHash []byte,
	latestCommitedBlockHash []byte,
	currentNonce uint64,
) *blockchainInfo {
	return &blockchainInfo{
		latestExecutedBlockHash: latestExecutedBlockHash,
		latestCommitedBlockHash: latestCommitedBlockHash,
		currentNonce:            currentNonce,
	}
}

// GetLatestExecutedBlockHash returns the hash of the latest executed block on blockchain
func (b *blockchainInfo) GetLatestExecutedBlockHash() []byte {
	return b.latestExecutedBlockHash
}

// GetLatestCommitedBlockHash returns the hash of the latest commited block on blockchain
func (b *blockchainInfo) GetLatestCommitedBlockHash() []byte {
	return b.latestCommitedBlockHash
}

// GetCurrentNonce returns the current nonce on blockchain
func (b *blockchainInfo) GetCurrentNonce() uint64 {
	return b.currentNonce
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *blockchainInfo) IsInterfaceNil() bool {
	return b == nil
}
