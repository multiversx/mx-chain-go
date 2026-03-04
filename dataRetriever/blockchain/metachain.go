package blockchain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

var _ data.ChainHandler = (*metaChain)(nil)

// The MetaChain also holds pointers to the Genesis block and the current block.
type metaChain struct {
	*baseBlockChain
	currentBlockRootHash []byte
}

// NewMetaChain will initialize a new metachain instance
func NewMetaChain(appStatusHandler core.AppStatusHandler) (*metaChain, error) {
	if check.IfNil(appStatusHandler) {
		return nil, ErrNilAppStatusHandler
	}

	return &metaChain{
		baseBlockChain: &baseBlockChain{
			appStatusHandler:      appStatusHandler,
			finalBlockInfo:        &blockInfo{},
			lastExecutedBlockInfo: &blockInfo{},
		},
	}, nil
}

// SetGenesisHeader returns the genesis block header pointer
func (mc *metaChain) SetGenesisHeader(header data.HeaderHandler) error {
	if check.IfNil(header) {
		mc.mut.Lock()
		mc.genesisHeader = nil
		mc.mut.Unlock()

		return nil
	}

	genBlock, ok := header.(data.MetaHeaderHandler)
	if !ok {
		return ErrWrongTypeInSet
	}
	mc.mut.Lock()
	mc.genesisHeader = genBlock.ShallowClone()
	mc.mut.Unlock()

	return nil
}

// SetCurrentBlockHeader sets current block header pointer
func (mc *metaChain) SetCurrentBlockHeader(header data.HeaderHandler) error {
	mc.mut.Lock()
	defer mc.mut.Unlock()

	return mc.setCurrentBlockHeaderUnprotected(header)
}

func (mc *metaChain) setCurrentBlockHeaderUnprotected(header data.HeaderHandler) error {
	if check.IfNil(header) {
		mc.currentBlockHeader = nil
		return nil
	}

	currHead, ok := header.(data.MetaHeaderHandler)
	if !ok {
		return ErrWrongTypeInSet
	}

	mc.currentBlockHeader = currHead.ShallowClone()

	mc.setCurrentHeaderMetrics(header)

	return nil
}

// SetCurrentBlockHeaderAndRootHash sets current block header pointer and the root hash
func (mc *metaChain) SetCurrentBlockHeaderAndRootHash(header data.HeaderHandler, rootHash []byte) error {
	mc.mut.Lock()
	defer mc.mut.Unlock()

	err := mc.setCurrentBlockHeaderUnprotected(header)
	if err != nil {
		return err
	}

	mc.currentBlockRootHash = make([]byte, len(rootHash))
	copy(mc.currentBlockRootHash, rootHash)

	return nil
}

// GetCurrentBlockRootHash returns the current committed block root hash. The returned byte slice is a new copy
// of the contained root hash.
func (mc *metaChain) GetCurrentBlockRootHash() []byte {
	mc.mut.RLock()
	rootHash := mc.currentBlockRootHash
	mc.mut.RUnlock()

	cloned := make([]byte, len(rootHash))
	copy(cloned, rootHash)

	return cloned
}

// IsInterfaceNil returns true if there is no value under the interface
func (mc *metaChain) IsInterfaceNil() bool {
	return mc == nil || mc.baseBlockChain == nil
}
