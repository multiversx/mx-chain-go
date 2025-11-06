package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ChainHandlerMock -
type ChainHandlerMock struct {
	genesisBlockHeader data.HeaderHandler
	genesisBlockHash   []byte

	currentBlockHeader   data.HeaderHandler
	currentBlockHash     []byte
	currentBlockRootHash []byte

	finalBlockNonce    uint64
	finalBlockHash     []byte
	finalBlockRootHash []byte

	lastExecutedBlockNonce    uint64
	lastExecutedBlockHash     []byte
	lastExecutedBlockRootHash []byte
	lastExecutedBlockHeader   data.HeaderHandler
}

// GetGenesisHeader -
func (mock *ChainHandlerMock) GetGenesisHeader() data.HeaderHandler {
	return mock.genesisBlockHeader
}

// SetGenesisHeader -
func (mock *ChainHandlerMock) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	mock.genesisBlockHeader = genesisBlock
	return nil
}

// GetGenesisHeaderHash -
func (mock *ChainHandlerMock) GetGenesisHeaderHash() []byte {
	return mock.genesisBlockHash
}

// SetGenesisHeaderHash -
func (mock *ChainHandlerMock) SetGenesisHeaderHash(hash []byte) {
	mock.genesisBlockHash = hash
}

// GetCurrentBlockHeader -
func (mock *ChainHandlerMock) GetCurrentBlockHeader() data.HeaderHandler {
	return mock.currentBlockHeader
}

// SetCurrentBlockHeaderAndRootHash -
func (mock *ChainHandlerMock) SetCurrentBlockHeaderAndRootHash(header data.HeaderHandler, rootHash []byte) error {
	mock.currentBlockHeader = header
	mock.currentBlockRootHash = rootHash
	return nil
}

// GetCurrentBlockHeaderHash -
func (mock *ChainHandlerMock) GetCurrentBlockHeaderHash() []byte {
	return mock.currentBlockHash
}

// SetCurrentBlockHeaderHash -
func (mock *ChainHandlerMock) SetCurrentBlockHeaderHash(hash []byte) {
	mock.currentBlockHash = hash
}

// GetCurrentBlockRootHash -
func (mock *ChainHandlerMock) GetCurrentBlockRootHash() []byte {
	return mock.currentBlockRootHash
}

// SetFinalBlockInfo -
func (mock *ChainHandlerMock) SetFinalBlockInfo(nonce uint64, headerHash []byte, rootHash []byte) {
	mock.finalBlockNonce = nonce
	mock.finalBlockHash = headerHash
	mock.finalBlockRootHash = rootHash
}

// SetLastExecutedBlockInfo -
func (mock *ChainHandlerMock) SetLastExecutedBlockInfo(nonce uint64, headerHash []byte, rootHash []byte) {
	mock.lastExecutedBlockNonce = nonce
	mock.lastExecutedBlockHash = headerHash
	mock.lastExecutedBlockRootHash = rootHash
}

// GetFinalBlockInfo -
func (mock *ChainHandlerMock) GetFinalBlockInfo() (nonce uint64, blockHash []byte, rootHash []byte) {
	return mock.finalBlockNonce, mock.finalBlockHash, mock.finalBlockRootHash
}

// GetLastExecutedBlockInfo -
func (mock *ChainHandlerMock) GetLastExecutedBlockInfo() (nonce uint64, blockHash []byte, rootHash []byte) {
	return mock.lastExecutedBlockNonce, mock.lastExecutedBlockHash, mock.lastExecutedBlockRootHash
}

// SetCurrentBlockHeader -
func (mock *ChainHandlerMock) SetCurrentBlockHeader(header data.HeaderHandler) error {
	mock.currentBlockHeader = header
	return nil
}

// GetLastExecutedBlockHeader -
func (mock *ChainHandlerMock) GetLastExecutedBlockHeader() data.HeaderHandler {
	return mock.lastExecutedBlockHeader
}

// SetLastExecutedBlockHeader -
func (mock *ChainHandlerMock) SetLastExecutedBlockHeader(header data.HeaderHandler) error {
	mock.lastExecutedBlockHeader = header
	return nil
}

// IsInterfaceNil -
func (mock *ChainHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
