package blockchain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
)

var _ data.ChainHandler = (*metaChain)(nil)

// metaChain holds the block information for the beacon chain
//
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
			appStatusHandler: appStatusHandler,
			finalBlockInfo:   &blockInfo{},
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

	genBlock, ok := header.(*block.MetaBlock)
	if !ok {
		return ErrWrongTypeInSet
	}
	mc.mut.Lock()
	mc.genesisHeader = genBlock.ShallowClone()
	mc.mut.Unlock()

	return nil
}

// SetCurrentBlockHeaderAndRootHash sets current block header pointer and the root hash
func (mc *metaChain) SetCurrentBlockHeaderAndRootHash(header data.HeaderHandler, rootHash []byte) error {
	if check.IfNil(header) {
		mc.mut.Lock()
		mc.currentBlockHeader = nil
		mc.currentBlockRootHash = nil
		mc.mut.Unlock()

		return nil
	}

	currHead, ok := header.(*block.MetaBlock)
	if !ok {
		return ErrWrongTypeInSet
	}

	mc.appStatusHandler.SetUInt64Value(common.MetricNonce, currHead.Nonce)
	mc.appStatusHandler.SetUInt64Value(common.MetricSynchronizedRound, currHead.Round)
	mc.appStatusHandler.SetUInt64Value(common.MetricBlockTimestamp, currHead.GetTimeStamp())

	mc.mut.Lock()
	mc.currentBlockHeader = currHead.ShallowClone()
	mc.currentBlockRootHash = make([]byte, len(rootHash))
	copy(mc.currentBlockRootHash, rootHash)
	mc.mut.Unlock()

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
