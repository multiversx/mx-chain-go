package preprocess

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.DataMarshalizer = (*validatorInfoPreprocessor)(nil)
var _ process.PreProcessor = (*validatorInfoPreprocessor)(nil)

type validatorInfoPreprocessor struct {
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	blockSizeComputation BlockSizeComputationHandler
}

// NewValidatorInfoPreprocessor creates a new validatorInfo preprocessor object
func NewValidatorInfoPreprocessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockSizeComputation BlockSizeComputationHandler,
) (*validatorInfoPreprocessor, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(blockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
	}

	rtp := &validatorInfoPreprocessor{
		hasher:               hasher,
		marshalizer:          marshalizer,
		blockSizeComputation: blockSizeComputation,
	}
	return rtp, nil
}

// IsDataPrepared does nothing
func (vip *validatorInfoPreprocessor) IsDataPrepared(_ int, _ func() time.Duration) error {
	return nil
}

// RemoveTxBlockFromPools removes the peerMiniBlock from pool
func (vip *validatorInfoPreprocessor) RemoveTxBlockFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}
	if check.IfNil(miniBlockPool) {
		return process.ErrNilMiniBlockPool
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		currentMiniBlock := body.MiniBlocks[i]
		if currentMiniBlock.Type != block.PeerBlock {
			continue
		}

		miniBlockHash, err := core.CalculateHash(vip.marshalizer, vip.hasher, currentMiniBlock)
		if err != nil {
			return err
		}

		miniBlockPool.Remove(miniBlockHash)
	}

	return nil
}

// RestoreTxBlockIntoPools restores the peerMiniBlock to the pool
func (vip *validatorInfoPreprocessor) RestoreTxBlockIntoPools(
	body *block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	if check.IfNil(body) {
		return 0, process.ErrNilBlockBody
	}
	if check.IfNil(miniBlockPool) {
		return 0, process.ErrNilMiniBlockPool
	}

	validatorsInfoRestored := 0
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		miniBlockHash, err := core.CalculateHash(vip.marshalizer, vip.hasher, miniBlock)
		if err != nil {
			return validatorsInfoRestored, err
		}

		miniBlockPool.Put(miniBlockHash, miniBlock, miniBlock.Size())

		validatorsInfoRestored += len(miniBlock.TxHashes)
	}

	return validatorsInfoRestored, nil
}

// ProcessBlockTransactions does nothing
func (vip *validatorInfoPreprocessor) ProcessBlockTransactions(
	_ *block.Body,
	_ func() bool,
) error {
	return nil
}

// SaveTxBlockToStorage does nothing
func (vip *validatorInfoPreprocessor) SaveTxBlockToStorage(_ *block.Body) error {
	return nil
}

// CreateBlockStarted does nothing
func (vip *validatorInfoPreprocessor) CreateBlockStarted() {
}

// RequestBlockTransactions does nothing
func (vip *validatorInfoPreprocessor) RequestBlockTransactions(_ *block.Body) int {
	return 0
}

// RequestTransactionsForMiniBlock does nothing
func (vip *validatorInfoPreprocessor) RequestTransactionsForMiniBlock(_ *block.MiniBlock) int {
	return 0
}

// CreateAndProcessMiniBlocks does nothing
func (vip *validatorInfoPreprocessor) CreateAndProcessMiniBlocks(_ func() bool) (block.MiniBlockSlice, error) {
	// validatorsInfo are created only by meta
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock does nothing
func (vip *validatorInfoPreprocessor) ProcessMiniBlock(miniBlock *block.MiniBlock, _ func() bool, _ func() (int, int)) ([][]byte, error) {
	if miniBlock.Type != block.PeerBlock {
		return nil, process.ErrWrongTypeInMiniBlock
	}
	if miniBlock.SenderShardID != core.MetachainShardId {
		return nil, process.ErrValidatorInfoMiniBlockNotFromMeta
	}

	//TODO: We need another function in the BlockSizeComputationHandler implementation that will better handle
	//the PeerBlock miniblocks as those are not hashes
	if vip.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlock.TxHashes)) {
		return nil, process.ErrMaxBlockSizeReached
	}

	vip.blockSizeComputation.AddNumMiniBlocks(1)
	vip.blockSizeComputation.AddNumTxs(len(miniBlock.TxHashes))

	return nil, nil
}

// CreateMarshalizedData does nothing
func (vip *validatorInfoPreprocessor) CreateMarshalizedData(_ [][]byte) ([][]byte, error) {
	marshalized := make([][]byte, 0)
	return marshalized, nil
}

// GetAllCurrentUsedTxs does nothing
func (vip *validatorInfoPreprocessor) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	validatorInfoTxPool := make(map[string]data.TransactionHandler)
	return validatorInfoTxPool
}

// IsInterfaceNil does nothing
func (vip *validatorInfoPreprocessor) IsInterfaceNil() bool {
	return vip == nil
}
