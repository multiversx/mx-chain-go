package blockchain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

var _ data.ChainHandler = (*blockChain)(nil)

// blockChain holds the block information for the current shard.
//
// The BlockChain also holds pointers to the Genesis block header and the current block
type blockChain struct {
	*baseBlockChain
	currentBlockRootHash []byte
}

// NewBlockChain returns an initialized blockchain
func NewBlockChain(appStatusHandler core.AppStatusHandler) (*blockChain, error) {
	if check.IfNil(appStatusHandler) {
		return nil, ErrNilAppStatusHandler
	}
	return &blockChain{
		baseBlockChain: &baseBlockChain{
			appStatusHandler:      appStatusHandler,
			finalBlockInfo:        &blockInfo{},
			lastExecutedBlockInfo: &blockInfo{},
		},
	}, nil
}

// SetGenesisHeader sets the genesis block header pointer
func (bc *blockChain) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if check.IfNil(genesisBlock) {
		bc.mut.Lock()
		bc.genesisHeader = nil
		bc.mut.Unlock()

		return nil
	}

	gb, ok := genesisBlock.(data.ShardHeaderHandler)
	if !ok {
		return data.ErrInvalidHeaderType
	}
	bc.mut.Lock()
	bc.genesisHeader = gb.ShallowClone()
	bc.mut.Unlock()

	return nil
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *blockChain) SetCurrentBlockHeader(header data.HeaderHandler) error {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	return bc.setCurrentBlockHeaderUnprotected(header)
}

func (bc *blockChain) setCurrentBlockHeaderUnprotected(header data.HeaderHandler) error {
	if check.IfNil(header) {
		bc.currentBlockHeader = nil
		return nil
	}

	h, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return data.ErrInvalidHeaderType
	}

	bc.currentBlockHeader = h.ShallowClone()

	bc.setCurrentHeaderMetrics(header)

	return nil
}

// SetCurrentBlockHeaderAndRootHash sets current block header pointer and the root hash
func (bc *blockChain) SetCurrentBlockHeaderAndRootHash(header data.HeaderHandler, rootHash []byte) error {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	err := bc.setCurrentBlockHeaderUnprotected(header)
	if err != nil {
		return err
	}

	bc.currentBlockRootHash = make([]byte, len(rootHash))
	copy(bc.currentBlockRootHash, rootHash)

	return nil
}

// GetCurrentBlockRootHash returns the current committed block root hash. The returned byte slice is a new copy
// of the contained root hash.
func (bc *blockChain) GetCurrentBlockRootHash() []byte {
	bc.mut.RLock()
	rootHash := bc.currentBlockRootHash
	bc.mut.RUnlock()

	cloned := make([]byte, len(rootHash))
	copy(cloned, rootHash)

	return cloned
}

// IsInterfaceNil returns true if there is no value under the interface
func (bc *blockChain) IsInterfaceNil() bool {
	return bc == nil || bc.baseBlockChain == nil
}
