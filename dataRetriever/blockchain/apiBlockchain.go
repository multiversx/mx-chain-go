package blockchain

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type apiBlockchain struct {
	mainBlockchain  data.ChainHandler
	currentRootHash []byte
	mut             sync.RWMutex
}

// NewApiBlockchain returns a new instance of apiBlockchain
func NewApiBlockchain(mainBlockchain data.ChainHandler) (*apiBlockchain, error) {
	if check.IfNil(mainBlockchain) {
		return nil, ErrNilBlockChain
	}

	return &apiBlockchain{
		mainBlockchain: mainBlockchain,
	}, nil
}

// GetGenesisHeader returns the genesis header from the main chain handler
func (abc *apiBlockchain) GetGenesisHeader() data.HeaderHandler {
	return abc.mainBlockchain.GetGenesisHeader()
}

// SetGenesisHeader returns an error
func (abc *apiBlockchain) SetGenesisHeader(_ data.HeaderHandler) error {
	return ErrOperationNotPermitted
}

// GetGenesisHeaderHash returns the genesis header hash from the main chain handler
func (abc *apiBlockchain) GetGenesisHeaderHash() []byte {
	return abc.mainBlockchain.GetGenesisHeaderHash()
}

// SetGenesisHeaderHash does nothing
func (abc *apiBlockchain) SetGenesisHeaderHash(_ []byte) {
}

// GetCurrentBlockHeader returns the current block header from the main chain handler
func (abc *apiBlockchain) GetCurrentBlockHeader() data.HeaderHandler {
	return abc.mainBlockchain.GetCurrentBlockHeader()
}

// SetCurrentBlockHeaderAndRootHash saves the root hash locally, which will be returned until a further call with an empty rootHash is made
func (abc *apiBlockchain) SetCurrentBlockHeaderAndRootHash(_ data.HeaderHandler, rootHash []byte) error {
	abc.mut.Lock()
	abc.currentRootHash = rootHash
	abc.mut.Unlock()

	return nil
}

// GetCurrentBlockHeaderHash returns the current block header hash from the main chain handler
func (abc *apiBlockchain) GetCurrentBlockHeaderHash() []byte {
	return abc.mainBlockchain.GetCurrentBlockHeaderHash()
}

// SetCurrentBlockHeaderHash does nothing
func (abc *apiBlockchain) SetCurrentBlockHeaderHash(_ []byte) {
}

// GetCurrentBlockRootHash returns the current block root hash from the main chain handler, if no local root hash is set
// if there is a local root hash, it will be returned until its reset
func (abc *apiBlockchain) GetCurrentBlockRootHash() []byte {
	abc.mut.RLock()
	if len(abc.currentRootHash) > 0 {
		defer abc.mut.RUnlock()
		return abc.currentRootHash
	}

	abc.mut.RUnlock()

	return abc.mainBlockchain.GetCurrentBlockRootHash()
}

// SetFinalBlockInfo does nothing
func (abc *apiBlockchain) SetFinalBlockInfo(_ uint64, _ []byte, _ []byte) {
}

// GetFinalBlockInfo returns the final block header hash from the main chain handler
func (abc *apiBlockchain) GetFinalBlockInfo() (nonce uint64, blockHash []byte, rootHash []byte) {
	return abc.mainBlockchain.GetFinalBlockInfo()
}

// IsInterfaceNil returns true if there is no value under the interface
func (abc *apiBlockchain) IsInterfaceNil() bool {
	return abc == nil
}
