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
	lastExecutedResult        data.BaseExecutionResultHandler
}

// GetLastExecutionResult -
func (mock *ChainHandlerMock) GetLastExecutionResult() data.BaseExecutionResultHandler {
	return mock.lastExecutedResult
}

// SetLastExecutionResult -
func (mock *ChainHandlerMock) SetLastExecutionInfo(header data.HeaderHandler, result data.BaseExecutionResultHandler) {
	mock.lastExecutedResult = result
	mock.lastExecutedBlockHeader = header
	mock.lastExecutedBlockNonce = header.GetNonce()
	mock.lastExecutedBlockHash = result.GetHeaderHash()
	mock.lastExecutedBlockRootHash = result.GetRootHash()
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

// GetCurrentBlockHeaderAndHash -
func (mock *ChainHandlerMock) GetCurrentBlockHeaderAndHash() (data.HeaderHandler, []byte) {
	return mock.currentBlockHeader, mock.currentBlockHash
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

// SetCurrentBlockHeaderAndHash -
func (mock *ChainHandlerMock) SetCurrentBlockHeaderAndHash(headerHash []byte, header data.HeaderHandler) error {
	mock.currentBlockHeader = header
	mock.currentBlockHash = headerHash
	return nil
}

// GetLastExecutedBlockHeader -
func (mock *ChainHandlerMock) GetLastExecutedBlockHeader() data.HeaderHandler {
	return mock.lastExecutedBlockHeader
}

// SetLastExecutedBlockHeaderAndRootHash -
func (mock *ChainHandlerMock) SetLastExecutedBlockHeaderAndRootHash(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
	mock.lastExecutedBlockHeader = header
	mock.lastExecutedBlockNonce = header.GetNonce()
	mock.lastExecutedBlockHash = blockHash
	mock.lastExecutedBlockRootHash = rootHash
}

// IsInterfaceNil -
func (mock *ChainHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
