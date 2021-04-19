package blockchain

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

var _ data.ChainHandler = (*metaChain)(nil)

// metaChain holds the block information for the beacon chain
//
// The MetaChain also holds pointers to the Genesis block and the current block.
type metaChain struct {
	*baseBlockChain
}

// NewMetaChain will initialize a new metachain instance
func NewMetaChain(appStatusHandler core.AppStatusHandler) (*metaChain, error) {
	if check.IfNil(appStatusHandler) {
		return nil, ErrNilAppStatusHandler
	}

	return &metaChain{
		baseBlockChain: &baseBlockChain{
			appStatusHandler: appStatusHandler,
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

// SetCurrentBlockHeader sets current block header pointer
func (mc *metaChain) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if check.IfNil(header) {
		mc.mut.Lock()
		mc.currentBlockHeader = nil
		mc.mut.Unlock()

		return nil
	}

	currHead, ok := header.(*block.MetaBlock)
	if !ok {
		return ErrWrongTypeInSet
	}

	mc.appStatusHandler.SetUInt64Value(core.MetricNonce, currHead.Nonce)
	mc.appStatusHandler.SetUInt64Value(core.MetricSynchronizedRound, currHead.Round)

	mc.mut.Lock()
	mc.currentBlockHeader = currHead.ShallowClone()
	mc.mut.Unlock()

	return nil
}

// CreateNewHeader creates a new meta block
func (mc *metaChain) CreateNewHeader() data.HeaderHandler {
	return &block.MetaBlock{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mc *metaChain) IsInterfaceNil() bool {
	return mc == nil || mc.baseBlockChain == nil
}
