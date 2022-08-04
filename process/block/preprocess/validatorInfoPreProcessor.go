package preprocess

import (
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
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
	chReceivedAllValidatorsInfo        chan bool
	onRequestValidatorsInfo            func(txHashes [][]byte)
	validatorsInfoForBlock             txsForBlock
	validatorsInfoPool                 dataRetriever.ShardedDataCacherNotifier
	storage                            dataRetriever.StorageService
	refactorPeersMiniBlocksEnableEpoch uint32
	flagRefactorPeersMiniBlocks        atomic.Flag
}

// NewValidatorInfoPreprocessor creates a new validatorInfo preprocessor object
func NewValidatorInfoPreprocessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockSizeComputation BlockSizeComputationHandler,
	validatorsInfoPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	onRequestValidatorsInfo func(txHashes [][]byte),
	epochNotifier process.EpochNotifier,
	refactorPeersMiniBlocksEnableEpoch uint32,
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
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	bpp := &basePreProcess{
		hasher:               hasher,
		marshalizer:          marshalizer,
		blockSizeComputation: blockSizeComputation,
	}

	vip := &validatorInfoPreprocessor{
		basePreProcess:                     bpp,
		storage:                            store,
		validatorsInfoPool:                 validatorsInfoPool,
		onRequestValidatorsInfo:            onRequestValidatorsInfo,
		refactorPeersMiniBlocksEnableEpoch: refactorPeersMiniBlocksEnableEpoch,
	}

	vip.chReceivedAllValidatorsInfo = make(chan bool)
	vip.validatorsInfoPool.RegisterOnAdded(vip.receivedValidatorInfoTransaction)
	vip.validatorsInfoForBlock.txHashAndInfo = make(map[string]*txInfo)

	log.Debug("validatorInfoPreprocessor: enable epoch for refactor peers mini blocks", "epoch", vip.refactorPeersMiniBlocksEnableEpoch)

	epochNotifier.RegisterNotifyHandler(vip)

	return vip, nil
}

// waitForValidatorsInfoHashes waits for a call whether all the requested validators info appeared
//func (vip *validatorInfoPreprocessor) waitForValidatorsInfoHashes(waitTime time.Duration) error {
//	select {
//	case <-vip.chReceivedAllValidatorsInfo:
//		return nil
//	case <-time.After(waitTime):
//		return process.ErrTimeIsOut
//	}
//}

// IsDataPrepared returns non error if all the requested validators info arrived and were saved into the pool
func (vip *validatorInfoPreprocessor) IsDataPrepared(_ int, _ func() time.Duration) error {
	return nil
	//if requestedValidatorsInfo > 0 {
	//	log.Debug("requested missing validators info",
	//		"num validators info", requestedValidatorsInfo)
	//	err := vip.waitForValidatorsInfoHashes(haveTime())
	//	vip.validatorsInfoForBlock.mutTxsForBlock.Lock()
	//	missingValidatorsInfo := vip.validatorsInfoForBlock.missingTxs
	//	vip.validatorsInfoForBlock.missingTxs = 0
	//	vip.validatorsInfoForBlock.mutTxsForBlock.Unlock()
	//	log.Debug("received validators info",
	//		"num validators info", requestedValidatorsInfo-missingValidatorsInfo)
	//	if err != nil {
	//		return err
	//	}
	//}
	//return nil
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

		if vip.flagRefactorPeersMiniBlocks.IsSet() {
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

// SaveTxsToStorage saves the validators info from body into storage
func (vip *validatorInfoPreprocessor) SaveTxsToStorage(_ *block.Body) error {
	return nil
	//if check.IfNil(body) {
	//	return process.ErrNilBlockBody
	//}
	//
	//for i := 0; i < len(body.MiniBlocks); i++ {
	//	miniBlock := body.MiniBlocks[i]
	//	if miniBlock.Type != block.PeerBlock {
	//		continue
	//	}
	//
	//	vip.saveTxsToStorage(
	//		miniBlock.TxHashes,
	//		&vip.validatorsInfoForBlock,
	//		vip.storage,
	//		dataRetriever.UnsignedTransactionUnit,
	//	)
	//}
	//
	//return nil
}

// receivedValidatorInfoTransaction is a callback function called when a new validator info transaction
// is added in the validator info transactions pool
func (vip *validatorInfoPreprocessor) receivedValidatorInfoTransaction(_ []byte, value interface{}) {
	validatorInfo, ok := value.(*state.ShardValidatorInfo)
	if !ok {
		log.Warn("validatorInfoPreprocessor.receivedValidatorInfoTransaction", "error", process.ErrWrongTypeAssertion)
		return
	}

	log.Debug("validatorInfoPreprocessor.receivedValidatorInfoTransaction", "pk", validatorInfo.PublicKey)
	//receivedAllMissing := vip.baseReceivedTransaction(key, tx, &vip.validatorsInfoForBlock)
	//
	//if receivedAllMissing {
	//	vip.chReceivedAllValidatorsInfo <- true
	//}
}

// CreateBlockStarted cleans the local cache map for processed/created validators info at this round
func (vip *validatorInfoPreprocessor) CreateBlockStarted() {
	_ = core.EmptyChannel(vip.chReceivedAllValidatorsInfo)

	vip.validatorsInfoForBlock.mutTxsForBlock.Lock()
	vip.validatorsInfoForBlock.missingTxs = 0
	vip.validatorsInfoForBlock.txHashAndInfo = make(map[string]*txInfo)
	vip.validatorsInfoForBlock.mutTxsForBlock.Unlock()
}

// RequestBlockTransactions request for validators info if missing from a block.Body
func (vip *validatorInfoPreprocessor) RequestBlockTransactions(_ *block.Body) int {
	return 0
	//if check.IfNil(body) {
	//	return 0
	//}
	//
	//return vip.computeExistingAndRequestMissingValidatorsInfoForShards(body)
}

// computeExistingAndRequestMissingValidatorsInfoForShards calculates what validators info are available and requests
// what are missing from block.Body
//func (vip *validatorInfoPreprocessor) computeExistingAndRequestMissingValidatorsInfoForShards(body *block.Body) int {
//	validatorsInfoBody := block.Body{}
//	for _, mb := range body.MiniBlocks {
//		if mb.Type != block.PeerBlock {
//			continue
//		}
//		if mb.SenderShardID != core.MetachainShardId {
//			continue
//		}
//
//		validatorsInfoBody.MiniBlocks = append(validatorsInfoBody.MiniBlocks, mb)
//	}
//
//	numMissingTxsForShards := vip.computeExistingAndRequestMissing(
//		&validatorsInfoBody,
//		&vip.validatorsInfoForBlock,
//		vip.chReceivedAllValidatorsInfo,
//		vip.isMiniBlockCorrect,
//		vip.validatorsInfoPool,
//		vip.onRequestValidatorsInfoWithShard,
//	)
//
//	return numMissingTxsForShards
//}

//func (vip *validatorInfoPreprocessor) onRequestValidatorsInfoWithShard(_ uint32, txHashes [][]byte) {
//	vip.onRequestValidatorsInfo(txHashes)
//}

// RequestTransactionsForMiniBlock requests missing validators info for a certain miniblock
func (vip *validatorInfoPreprocessor) RequestTransactionsForMiniBlock(_ *block.MiniBlock) int {
	return 0
	//if miniBlock == nil {
	//	return 0
	//}
	//
	//missingValidatorsInfoHashesForMiniBlock := vip.computeMissingValidatorsInfoHashesForMiniBlock(miniBlock)
	//if len(missingValidatorsInfoHashesForMiniBlock) > 0 {
	//	vip.onRequestValidatorsInfo(missingValidatorsInfoHashesForMiniBlock)
	//}
	//
	//return len(missingValidatorsInfoHashesForMiniBlock)
}

// computeMissingValidatorsInfoHashesForMiniBlock computes missing validators info hashes for a certain miniblock
//func (vip *validatorInfoPreprocessor) computeMissingValidatorsInfoHashesForMiniBlock(miniBlock *block.MiniBlock) [][]byte {
//	missingValidatorsInfoHashes := make([][]byte, 0)
//	return missingValidatorsInfoHashes
//
//	if miniBlock.Type != block.PeerBlock {
//		return missingValidatorsInfoHashes
//	}
//
//	for _, txHash := range miniBlock.TxHashes {
//		validatorInfo, _ := process.GetValidatorInfoFromPool(
//			miniBlock.SenderShardID,
//			miniBlock.ReceiverShardID,
//			txHash,
//			vip.validatorsInfoPool,
//			false,
//		)
//
//		if validatorInfo == nil {
//			missingValidatorsInfoHashes = append(missingValidatorsInfoHashes, txHash)
//		}
//	}
//
//	return missingValidatorsInfoHashes
//}

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

// CreateMarshalledData marshals validators info hashes and saves them into a new structure
func (vip *validatorInfoPreprocessor) CreateMarshalledData(_ [][]byte) ([][]byte, error) {
	return make([][]byte, 0), nil
	//marshalledValidatorsInfo, err := vip.createMarshalledData(txHashes, &vip.validatorsInfoForBlock)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return marshalledValidatorsInfo, nil
}

// GetAllCurrentUsedTxs returns all the validators info used at current creation / processing
func (vip *validatorInfoPreprocessor) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	return make(map[string]data.TransactionHandler)
	//vip.validatorsInfoForBlock.mutTxsForBlock.RLock()
	//validatorsInfoPool := make(map[string]data.TransactionHandler, len(vip.validatorsInfoForBlock.txHashAndInfo))
	//for txHash, txData := range vip.validatorsInfoForBlock.txHashAndInfo {
	//	validatorsInfoPool[txHash] = txData.tx
	//}
	//vip.validatorsInfoForBlock.mutTxsForBlock.RUnlock()
	//
	//return validatorsInfoPool
}

// AddTxsFromMiniBlocks does nothing
func (vip *validatorInfoPreprocessor) AddTxsFromMiniBlocks(_ block.MiniBlockSlice) {
}

// AddTransactions does nothing
func (vip *validatorInfoPreprocessor) AddTransactions(_ []data.TransactionHandler) {
}

// IsInterfaceNil does nothing
func (vip *validatorInfoPreprocessor) IsInterfaceNil() bool {
	return vip == nil
}

func (vip *validatorInfoPreprocessor) isMiniBlockCorrect(mbType block.Type) bool {
	return mbType == block.PeerBlock
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (vip *validatorInfoPreprocessor) EpochConfirmed(epoch uint32, timestamp uint64) {
	vip.epochConfirmed(epoch, timestamp)

	vip.flagRefactorPeersMiniBlocks.SetValue(epoch >= vip.refactorPeersMiniBlocksEnableEpoch)
	log.Debug("validatorInfoPreprocessor: refactor peers mini blocks", "enabled", vip.flagRefactorPeersMiniBlocks.IsSet())
}
