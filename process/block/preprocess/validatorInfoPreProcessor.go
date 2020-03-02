package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type validatorInfoPreprocessor struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewValidatorInfoPreprocessor creates a new validatorInfo preprocessor object
func NewValidatorInfoPreprocessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) (*validatorInfoPreprocessor, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	rtp := &validatorInfoPreprocessor{
		hasher:      hasher,
		marshalizer: marshalizer,
	}
	return rtp, nil
}

// IsDataPrepared returns nil because the TxHashes are the actual serialized objects neeed
func (rtp *validatorInfoPreprocessor) IsDataPrepared(_ int, _ func() time.Duration) error {
	return nil
}

// RemoveTxBlockFromPools removes reward transactions and miniblocks from associated pools
func (rtp *validatorInfoPreprocessor) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	return nil
}

// RestoreTxBlockIntoPools restores the reward transactions and miniblocks to associated pools
func (rtp *validatorInfoPreprocessor) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	if miniBlockPool == nil {
		return 0, process.ErrNilMiniBlockPool
	}

	rewardTxsRestored := 0
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		miniBlockHash, err := core.CalculateHash(rtp.marshalizer, rtp.hasher, miniBlock)
		if err != nil {
			return rewardTxsRestored, err
		}

		miniBlockPool.Put(miniBlockHash, miniBlock)

		rewardTxsRestored += len(miniBlock.TxHashes)
	}

	return rewardTxsRestored, nil

	return 0, nil
}

// ProcessBlockTransactions processes all the reward transactions from the block.Body, updates the state
func (rtp *validatorInfoPreprocessor) ProcessBlockTransactions(
	body block.Body,
	haveTime func() bool,
) error {
	return nil
}

// SaveTxBlockToStorage saves the reward transactions from body into storage
func (rtp *validatorInfoPreprocessor) SaveTxBlockToStorage(body block.Body) error {
	return nil
}

// CreateBlockStarted cleans the local cache map for processed/created reward transactions at this round
func (rtp *validatorInfoPreprocessor) CreateBlockStarted() {
}

// RequestBlockTransactions request for reward transactions if missing from a block.Body
func (rtp *validatorInfoPreprocessor) RequestBlockTransactions(body block.Body) int {
	return 0
}

// RequestTransactionsForMiniBlock requests missing reward transactions for a certain miniblock
func (rtp *validatorInfoPreprocessor) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	return 0
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (rtp *validatorInfoPreprocessor) CreateAndProcessMiniBlocks(
	_ uint32,
	_ uint32,
	_ func() bool,
) (block.MiniBlockSlice, error) {
	// rewards are created only by meta
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock processes all the reward transactions from a miniblock and saves the processed reward transactions
// in local cache
func (rtp *validatorInfoPreprocessor) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool) error {
	if miniBlock.Type != block.PeerBlock {
		return process.ErrWrongTypeInMiniBlock
	}
	if miniBlock.SenderShardID != core.MetachainShardId {
		return process.ErrRewardMiniBlockNotFromMeta
	}

	return nil
}

// CreateMarshalizedData marshalizes reward transaction hashes and and saves them into a new structure
func (rtp *validatorInfoPreprocessor) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	marshalized := make([][]byte, 0)
	return marshalized, nil
}

// GetAllCurrentUsedTxs returns all the reward transactions used at current creation / processing
func (rtp *validatorInfoPreprocessor) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	rewardTxPool := make(map[string]data.TransactionHandler, 0)

	return rewardTxPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *validatorInfoPreprocessor) IsInterfaceNil() bool {
	return rtp == nil
}
