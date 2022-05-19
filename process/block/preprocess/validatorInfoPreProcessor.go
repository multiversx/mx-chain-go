package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.DataMarshalizer = (*validatorInfoPreprocessor)(nil)
var _ process.PreProcessor = (*validatorInfoPreprocessor)(nil)

type validatorInfoPreprocessor struct {
	*basePreProcess
	chReceivedAllValidatorInfo chan bool
	onRequestValidatorsInfo    func(txHashes [][]byte)
	validatorInfoForBlock      txsForBlock
	validatorInfoPool          dataRetriever.ShardedDataCacherNotifier
	storage                    dataRetriever.StorageService
}

// NewValidatorInfoPreprocessor creates a new validatorInfo preprocessor object
func NewValidatorInfoPreprocessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockSizeComputation BlockSizeComputationHandler,
	validatorInfoPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	onRequestValidatorsInfo func(txHashes [][]byte),
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
	if check.IfNil(validatorInfoPool) {
		return nil, process.ErrNilValidatorInfoPool
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}
	if onRequestValidatorsInfo == nil {
		return nil, process.ErrNilRequestHandler
	}

	bpp := &basePreProcess{
		hasher:               hasher,
		marshalizer:          marshalizer,
		blockSizeComputation: blockSizeComputation,
	}

	vip := &validatorInfoPreprocessor{
		basePreProcess:          bpp,
		storage:                 store,
		validatorInfoPool:       validatorInfoPool,
		onRequestValidatorsInfo: onRequestValidatorsInfo,
	}

	vip.chReceivedAllValidatorInfo = make(chan bool)
	vip.validatorInfoPool.RegisterOnAdded(vip.receivedValidatorInfoTransaction)
	vip.validatorInfoForBlock.txHashAndInfo = make(map[string]*txInfo)

	return vip, nil
}

// IsDataPrepared does nothing
func (vip *validatorInfoPreprocessor) IsDataPrepared(_ int, _ func() time.Duration) error {
	return nil
}

// RemoveBlockDataFromPools removes the peer miniblocks from pool
func (vip *validatorInfoPreprocessor) RemoveBlockDataFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	return vip.removeBlockDataFromPools(body, miniBlockPool, vip.validatorInfoPool, vip.isMiniBlockCorrect)
}

// RemoveTxsFromPools does nothing for validatorInfoPreprocessor implementation
func (vip *validatorInfoPreprocessor) RemoveTxsFromPools(_ *block.Body) error {
	return nil
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
func (vip *validatorInfoPreprocessor) receivedValidatorInfoTransaction(key []byte, value interface{}) {
	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("validatorInfoPreprocessor.receivedValidatorInfoTransaction", "error", process.ErrWrongTypeAssertion)
		return
	}

	receivedAllMissing := vip.baseReceivedTransaction(key, tx, &vip.validatorInfoForBlock)

	if receivedAllMissing {
		vip.chReceivedAllValidatorInfo <- true
	}
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
func (vip *validatorInfoPreprocessor) CreateAndProcessMiniBlocks(_ func() bool, _ []byte) (block.MiniBlockSlice, error) {
	// validatorsInfo are created only by meta
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock does nothing
func (vip *validatorInfoPreprocessor) ProcessMiniBlock(miniBlock *block.MiniBlock, _ func() bool, _ func() bool, _ func() (int, int), _ bool) ([][]byte, int, error) {
	if miniBlock.Type != block.PeerBlock {
		return nil, 0, process.ErrWrongTypeInMiniBlock
	}
	if miniBlock.SenderShardID != core.MetachainShardId {
		return nil, 0, process.ErrValidatorInfoMiniBlockNotFromMeta
	}

	//TODO: We need another function in the BlockSizeComputationHandler implementation that will better handle
	//the PeerBlock miniblocks as those are not hashes
	if vip.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlock.TxHashes)) {
		return nil, 0, process.ErrMaxBlockSizeReached
	}

	vip.blockSizeComputation.AddNumMiniBlocks(1)
	vip.blockSizeComputation.AddNumTxs(len(miniBlock.TxHashes))

	return nil, len(miniBlock.TxHashes), nil
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
