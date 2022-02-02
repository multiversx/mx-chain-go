package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
)

var _ data.ChainHandler = (*blockChain)(nil)

// blockChain holds the block information for the current shard.
//
// The BlockChain also holds pointers to the Genesis block header and the current block
type blockChain struct {
	*baseBlockChain
}

// NewBlockChain returns an initialized blockchain
func NewBlockChain(appStatusHandler core.AppStatusHandler) (*blockChain, error) {
	if check.IfNil(appStatusHandler) {
		return nil, ErrNilAppStatusHandler
	}
	return &blockChain{
		baseBlockChain: &baseBlockChain{
			appStatusHandler: appStatusHandler,
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
	if check.IfNil(header) {
		bc.mut.Lock()
		bc.currentBlockHeader = nil
		bc.mut.Unlock()

		return nil
	}

	h, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return data.ErrInvalidHeaderType
	}

	bc.appStatusHandler.SetUInt64Value(common.MetricNonce, h.GetNonce())
	bc.appStatusHandler.SetUInt64Value(common.MetricSynchronizedRound, h.GetRound())

	bc.mut.Lock()
	bc.currentBlockHeader = h.ShallowClone()
	bc.mut.Unlock()

	return nil
}

// GetCurrentBlockCommittedRootHash returns the current committed block root hash. If the schedule root hash is available,
// it will return that value, otherwise the block's root hash. If there is no current block set, it will return nil
func (bc *blockChain) GetCurrentBlockCommittedRootHash() []byte {
	bc.mut.RLock()
	currHead := bc.currentBlockHeader
	bc.mut.RUnlock()

	if check.IfNil(currHead) {
		return nil
	}
	rootHash := currHead.GetRootHash()
	additionalData := currHead.GetAdditionalData()
	if additionalData == nil {
		return rootHash
	}
	scheduledRootHash := additionalData.GetScheduledRootHash()
	if len(scheduledRootHash) == 0 {
		return rootHash
	}

	return scheduledRootHash
}

// IsInterfaceNil returns true if there is no value under the interface
func (bc *blockChain) IsInterfaceNil() bool {
	return bc == nil || bc.baseBlockChain == nil
}
