package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.DataMarshalizer = (*validatorInfoPreprocessor)(nil)
var _ process.PreProcessor = (*validatorInfoPreprocessor)(nil)

type validatorInfoPreprocessor struct {
	*basePreProcess
	chReceivedAllValidatorsInfo chan bool
	onRequestValidatorsInfo     func(txHashes [][]byte)
	validatorsInfoForBlock      txsForBlock
	validatorsInfoPool          dataRetriever.ShardedDataCacherNotifier
	storage                     dataRetriever.StorageService
	enableEpochsHandler         common.EnableEpochsHandler
}

// NewValidatorInfoPreprocessor creates a new validatorInfo preprocessor object
func NewValidatorInfoPreprocessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockSizeComputation BlockSizeComputationHandler,
	validatorsInfoPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	onRequestValidatorsInfo func(txHashes [][]byte),
	enableEpochsHandler common.EnableEpochsHandler,
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
	if check.IfNil(validatorsInfoPool) {
		return nil, process.ErrNilValidatorInfoPool
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}
	if onRequestValidatorsInfo == nil {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	bpp := &basePreProcess{
		hasher:               hasher,
		marshalizer:          marshalizer,
		blockSizeComputation: blockSizeComputation,
	}

	vip := &validatorInfoPreprocessor{
		basePreProcess:          bpp,
		storage:                 store,
		validatorsInfoPool:      validatorsInfoPool,
		onRequestValidatorsInfo: onRequestValidatorsInfo,
		enableEpochsHandler:     enableEpochsHandler,
	}

	vip.chReceivedAllValidatorsInfo = make(chan bool)
	vip.validatorsInfoPool.RegisterOnAdded(vip.receivedValidatorInfoTransaction)
	vip.validatorsInfoForBlock.txHashAndInfo = make(map[string]*txInfo)

	return vip, nil
}

// IsDataPrepared returns non error if all the requested validators info arrived and were saved into the pool
func (vip *validatorInfoPreprocessor) IsDataPrepared(_ int, _ func() time.Duration) error {
	return nil
}

// RemoveBlockDataFromPools removes the peer miniblocks from pool
func (vip *validatorInfoPreprocessor) RemoveBlockDataFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	return vip.removeBlockDataFromPools(body, miniBlockPool, vip.validatorsInfoPool, vip.isMiniBlockCorrect)
}

// RemoveTxsFromPools removes validators info from associated pools
func (vip *validatorInfoPreprocessor) RemoveTxsFromPools(body *block.Body) error {
	return vip.removeTxsFromPools(body, vip.validatorsInfoPool, vip.isMiniBlockCorrect)
}

// RestoreBlockDataIntoPools restores the peer miniblocks to the pool
func (vip *validatorInfoPreprocessor) RestoreBlockDataIntoPools(
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

		if vip.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
			err := vip.restoreValidatorsInfo(miniBlock)
			if err != nil {
				return validatorsInfoRestored, err
			}
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

func (vip *validatorInfoPreprocessor) restoreValidatorsInfo(miniBlock *block.MiniBlock) error {
	strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
	validatorsInfoBuff, err := vip.storage.GetAll(dataRetriever.UnsignedTransactionUnit, miniBlock.TxHashes)
	if err != nil {
		log.Debug("validators info from mini block were not found in UnsignedTransactionUnit",
			"sender shard ID", miniBlock.SenderShardID,
			"receiver shard ID", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes),
		)

		return err
	}

	for validatorInfoHash, validatorInfoBuff := range validatorsInfoBuff {
		shardValidatorInfo := &state.ShardValidatorInfo{}
		err = vip.marshalizer.Unmarshal(shardValidatorInfo, validatorInfoBuff)
		if err != nil {
			return err
		}

		vip.validatorsInfoPool.AddData([]byte(validatorInfoHash), shardValidatorInfo, shardValidatorInfo.Size(), strCache)
	}

	return nil
}

// ProcessBlockTransactions does nothing
func (vip *validatorInfoPreprocessor) ProcessBlockTransactions(
	_ data.HeaderHandler,
	_ *block.Body,
	_ func() bool,
) error {
	return nil
}

// SaveTxsToStorage does nothing
func (vip *validatorInfoPreprocessor) SaveTxsToStorage(_ *block.Body) error {
	return nil
}

// receivedValidatorInfoTransaction is a callback function called when a new validator info transaction
// is added in the validator info transactions pool
func (vip *validatorInfoPreprocessor) receivedValidatorInfoTransaction(txHash []byte, value interface{}) {
	validatorInfo, ok := value.(*state.ShardValidatorInfo)
	if !ok {
		log.Warn("validatorInfoPreprocessor.receivedValidatorInfoTransaction", "error", process.ErrWrongTypeAssertion)
		return
	}

	log.Trace("validatorInfoPreprocessor.receivedValidatorInfoTransaction", "tx hash", txHash, "pk", validatorInfo.PublicKey)
}

// CreateBlockStarted cleans the local cache map for processed/created validators info at this round
func (vip *validatorInfoPreprocessor) CreateBlockStarted() {
	_ = core.EmptyChannel(vip.chReceivedAllValidatorsInfo)

	vip.validatorsInfoForBlock.mutTxsForBlock.Lock()
	vip.validatorsInfoForBlock.missingTxs = 0
	vip.validatorsInfoForBlock.txHashAndInfo = make(map[string]*txInfo)
	vip.validatorsInfoForBlock.mutTxsForBlock.Unlock()
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
func (vip *validatorInfoPreprocessor) CreateAndProcessMiniBlocks(_ func() bool, _ []byte) (block.MiniBlockSlice, error) {
	// validatorsInfo are created only by meta
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock does nothing
func (vip *validatorInfoPreprocessor) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	_ func() bool,
	_ func() bool,
	_ bool,
	_ bool,
	indexOfLastTxProcessed int,
	_ process.PreProcessorExecutionInfoHandler,
) ([][]byte, int, bool, error) {
	if miniBlock.Type != block.PeerBlock {
		return nil, indexOfLastTxProcessed, false, process.ErrWrongTypeInMiniBlock
	}
	if miniBlock.SenderShardID != core.MetachainShardId {
		return nil, indexOfLastTxProcessed, false, process.ErrValidatorInfoMiniBlockNotFromMeta
	}

	if vip.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlock.TxHashes)) {
		return nil, indexOfLastTxProcessed, false, process.ErrMaxBlockSizeReached
	}

	vip.blockSizeComputation.AddNumMiniBlocks(1)
	vip.blockSizeComputation.AddNumTxs(len(miniBlock.TxHashes))

	return nil, len(miniBlock.TxHashes) - 1, false, nil
}

// CreateMarshalledData does nothing
func (vip *validatorInfoPreprocessor) CreateMarshalledData(_ [][]byte) ([][]byte, error) {
	return make([][]byte, 0), nil
}

// GetAllCurrentUsedTxs does nothing
func (vip *validatorInfoPreprocessor) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	return make(map[string]data.TransactionHandler)
}

// AddTxsFromMiniBlocks does nothing
func (vip *validatorInfoPreprocessor) AddTxsFromMiniBlocks(_ block.MiniBlockSlice) {
}

// AddTransactions does nothing
func (vip *validatorInfoPreprocessor) AddTransactions(_ []data.TransactionHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (vip *validatorInfoPreprocessor) IsInterfaceNil() bool {
	return vip == nil
}

func (vip *validatorInfoPreprocessor) isMiniBlockCorrect(mbType block.Type) bool {
	return mbType == block.PeerBlock
}
