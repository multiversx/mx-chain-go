package preprocess

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

const initialAllocation = 1000

var _ process.GasHandler = (*gasComputation)(nil)

type gasComputation struct {
	economicsFee  process.FeeHandler
	txTypeHandler process.TxTypeHandler
	// TODO: Refactor these mutexes and maps in separated structures that handle the locking and unlocking for each operation required
	gasProvided                                      map[string]uint64
	txHashesWithGasProvidedSinceLastReset            map[string][][]byte
	gasProvidedAsScheduled                           map[string]uint64
	txHashesWithGasProvidedAsScheduledSinceLastReset map[string][][]byte
	mutGasProvided                                   sync.RWMutex
	gasRefunded                                      map[string]uint64
	txHashesWithGasRefundedSinceLastReset            map[string][][]byte
	mutGasRefunded                                   sync.RWMutex
	gasPenalized                                     map[string]uint64
	txHashesWithGasPenalizedSinceLastReset           map[string][][]byte
	mutGasPenalized                                  sync.RWMutex
	enableEpochsHandler                              common.EnableEpochsHandler
}

// NewGasComputation creates a new object which computes the gas consumption
func NewGasComputation(
	economicsFee process.FeeHandler,
	txTypeHandler process.TxTypeHandler,
	enableEpochsHandler common.EnableEpochsHandler,
) (*gasComputation, error) {
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	g := &gasComputation{
		txTypeHandler:                         txTypeHandler,
		economicsFee:                          economicsFee,
		gasProvided:                           make(map[string]uint64),
		txHashesWithGasProvidedSinceLastReset: make(map[string][][]byte),
		gasProvidedAsScheduled:                make(map[string]uint64),
		txHashesWithGasProvidedAsScheduledSinceLastReset: make(map[string][][]byte),
		gasRefunded:                            make(map[string]uint64),
		txHashesWithGasRefundedSinceLastReset:  make(map[string][][]byte),
		gasPenalized:                           make(map[string]uint64),
		txHashesWithGasPenalizedSinceLastReset: make(map[string][][]byte),
		enableEpochsHandler:                    enableEpochsHandler,
	}

	return g, nil
}

// Init method inits provided, refunded and penalized gas structures
func (gc *gasComputation) Init() {
	gc.mutGasProvided.Lock()
	gc.gasProvided = make(map[string]uint64)
	gc.txHashesWithGasProvidedSinceLastReset = make(map[string][][]byte)
	gc.gasProvidedAsScheduled = make(map[string]uint64)
	gc.txHashesWithGasProvidedAsScheduledSinceLastReset = make(map[string][][]byte)
	gc.mutGasProvided.Unlock()

	gc.mutGasRefunded.Lock()
	gc.gasRefunded = make(map[string]uint64)
	gc.txHashesWithGasRefundedSinceLastReset = make(map[string][][]byte)
	gc.mutGasRefunded.Unlock()

	gc.mutGasPenalized.Lock()
	gc.gasPenalized = make(map[string]uint64)
	gc.txHashesWithGasPenalizedSinceLastReset = make(map[string][][]byte)
	gc.mutGasPenalized.Unlock()
}

// Reset method resets tx hashes with gas provided, refunded and penalized since last reset
// TODO remove this call from basePreProcess.handleProcessTransactionInit
func (gc *gasComputation) Reset(key []byte) {
	gc.mutGasProvided.Lock()
	gc.txHashesWithGasProvidedSinceLastReset[string(key)] = make([][]byte, 0, initialAllocation)
	gc.txHashesWithGasProvidedAsScheduledSinceLastReset[string(key)] = make([][]byte, 0, initialAllocation)
	gc.mutGasProvided.Unlock()

	gc.mutGasRefunded.Lock()
	gc.txHashesWithGasRefundedSinceLastReset[string(key)] = make([][]byte, 0, initialAllocation)
	gc.mutGasRefunded.Unlock()

	gc.mutGasPenalized.Lock()
	gc.txHashesWithGasPenalizedSinceLastReset[string(key)] = make([][]byte, 0, initialAllocation)
	gc.mutGasPenalized.Unlock()
}

// SetGasProvided sets gas provided for a given hash
func (gc *gasComputation) SetGasProvided(gasProvided uint64, hash []byte) {
	gc.mutGasProvided.Lock()
	gc.gasProvided[string(hash)] = gasProvided
	for key := range gc.txHashesWithGasProvidedSinceLastReset {
		gc.txHashesWithGasProvidedSinceLastReset[key] = append(gc.txHashesWithGasProvidedSinceLastReset[key], hash)
	}
	gc.mutGasProvided.Unlock()
}

// SetGasProvidedAsScheduled sets gas provided as scheduled for a given hash
func (gc *gasComputation) SetGasProvidedAsScheduled(gasProvided uint64, hash []byte) {
	gc.mutGasProvided.Lock()
	gc.gasProvidedAsScheduled[string(hash)] = gasProvided
	for key := range gc.txHashesWithGasProvidedAsScheduledSinceLastReset {
		gc.txHashesWithGasProvidedAsScheduledSinceLastReset[key] = append(gc.txHashesWithGasProvidedAsScheduledSinceLastReset[key], hash)
	}
	gc.mutGasProvided.Unlock()
}

// SetGasRefunded sets gas refunded for a given hash
func (gc *gasComputation) SetGasRefunded(gasRefunded uint64, hash []byte) {
	gc.mutGasRefunded.Lock()
	gc.gasRefunded[string(hash)] = gasRefunded
	for key := range gc.txHashesWithGasRefundedSinceLastReset {
		gc.txHashesWithGasRefundedSinceLastReset[key] = append(gc.txHashesWithGasRefundedSinceLastReset[key], hash)
	}
	gc.mutGasRefunded.Unlock()

	log.Trace("gasComputation.SetGasRefunded", "tx hash", hash, "gas refunded", gasRefunded)
}

// SetGasPenalized sets gas penalized for a given hash
func (gc *gasComputation) SetGasPenalized(gasPenalized uint64, hash []byte) {
	gc.mutGasPenalized.Lock()
	gc.gasPenalized[string(hash)] = gasPenalized
	for key := range gc.txHashesWithGasPenalizedSinceLastReset {
		gc.txHashesWithGasPenalizedSinceLastReset[key] = append(gc.txHashesWithGasPenalizedSinceLastReset[key], hash)
	}
	gc.mutGasPenalized.Unlock()

	log.Trace("gasComputation.SetGasPenalized", "tx hash", hash, "gas penalized", gasPenalized)
}

// GasProvided gets gas provided for a given hash
func (gc *gasComputation) GasProvided(hash []byte) uint64 {
	gc.mutGasProvided.RLock()
	gasProvided := gc.gasProvided[string(hash)]
	gc.mutGasProvided.RUnlock()

	return gasProvided

}

// GasProvidedAsScheduled gets gas provided as scheduled for a given hash
func (gc *gasComputation) GasProvidedAsScheduled(hash []byte) uint64 {
	gc.mutGasProvided.RLock()
	gasProvidedAsScheduled := gc.gasProvidedAsScheduled[string(hash)]
	gc.mutGasProvided.RUnlock()

	return gasProvidedAsScheduled

}

// GasRefunded gets gas refunded for a given hash
func (gc *gasComputation) GasRefunded(hash []byte) uint64 {
	gc.mutGasRefunded.RLock()
	gasRefunded := gc.gasRefunded[string(hash)]
	gc.mutGasRefunded.RUnlock()

	return gasRefunded
}

// GasPenalized gets gas penalized for a given hash
func (gc *gasComputation) GasPenalized(hash []byte) uint64 {
	gc.mutGasPenalized.RLock()
	gasPenalized := gc.gasPenalized[string(hash)]
	gc.mutGasPenalized.RUnlock()

	return gasPenalized
}

// TotalGasProvided gets the total gas provided
func (gc *gasComputation) TotalGasProvided() uint64 {
	totalGasProvided := uint64(0)

	gc.mutGasProvided.RLock()
	for _, gasProvided := range gc.gasProvided {
		totalGasProvided += gasProvided
	}
	gc.mutGasProvided.RUnlock()

	return totalGasProvided
}

// TotalGasProvidedAsScheduled gets the total gas provided as scheduled
func (gc *gasComputation) TotalGasProvidedAsScheduled() uint64 {
	totalGasProvided := uint64(0)

	gc.mutGasProvided.RLock()
	for _, gasProvided := range gc.gasProvidedAsScheduled {
		totalGasProvided += gasProvided
	}
	gc.mutGasProvided.RUnlock()

	return totalGasProvided
}

// TotalGasProvidedWithScheduled gets the total gas provided both as normal and scheduled operation
func (gc *gasComputation) TotalGasProvidedWithScheduled() uint64 {
	gasProvidedAsScheduled := gc.TotalGasProvidedAsScheduled()
	gasProvidedNormalOperation := gc.TotalGasProvided()

	return gasProvidedAsScheduled + gasProvidedNormalOperation
}

// TotalGasRefunded gets the total gas refunded
func (gc *gasComputation) TotalGasRefunded() uint64 {
	totalGasRefunded := uint64(0)
	gc.mutGasRefunded.RLock()
	for _, gasRefunded := range gc.gasRefunded {
		totalGasRefunded += gasRefunded
	}
	gc.mutGasRefunded.RUnlock()

	return totalGasRefunded
}

// TotalGasPenalized gets the total gas penalized
func (gc *gasComputation) TotalGasPenalized() uint64 {
	totalGasPenalized := uint64(0)
	gc.mutGasPenalized.RLock()
	for _, gasPenalized := range gc.gasPenalized {
		totalGasPenalized += gasPenalized
	}
	gc.mutGasPenalized.RUnlock()

	return totalGasPenalized
}

// RemoveGasProvided removes gas provided for the given hashes
func (gc *gasComputation) RemoveGasProvided(hashes [][]byte) {
	gc.mutGasProvided.Lock()
	for _, hash := range hashes {
		delete(gc.gasProvided, string(hash))
	}
	gc.mutGasProvided.Unlock()
}

// RemoveGasProvidedAsScheduled removes gas provided as scheduled for the given hashes
func (gc *gasComputation) RemoveGasProvidedAsScheduled(hashes [][]byte) {
	gc.mutGasProvided.Lock()
	for _, hash := range hashes {
		delete(gc.gasProvidedAsScheduled, string(hash))
	}
	gc.mutGasProvided.Unlock()
}

// RemoveGasRefunded removes gas refunded for the given hashes
func (gc *gasComputation) RemoveGasRefunded(hashes [][]byte) {
	gc.mutGasRefunded.Lock()
	for _, hash := range hashes {
		delete(gc.gasRefunded, string(hash))
	}
	gc.mutGasRefunded.Unlock()
}

// RemoveGasPenalized removes gas penalized for the given hashes
func (gc *gasComputation) RemoveGasPenalized(hashes [][]byte) {
	gc.mutGasPenalized.Lock()
	for _, hash := range hashes {
		delete(gc.gasPenalized, string(hash))
	}
	gc.mutGasPenalized.Unlock()
}

// RestoreGasSinceLastReset method restores gas provided, refunded and penalized since last reset
func (gc *gasComputation) RestoreGasSinceLastReset(key []byte) {
	gc.RemoveGasProvided(gc.getTxHashesWithGasProvidedSinceLastReset(key))
	gc.RemoveGasProvidedAsScheduled(gc.getTxHashesWithGasProvidedAsScheduledSinceLastReset(key))
	gc.RemoveGasRefunded(gc.getTxHashesWithGasRefundedSinceLastReset(key))
	gc.RemoveGasPenalized(gc.getTxHashesWithGasPenalizedSinceLastReset(key))
}

func (gc *gasComputation) getTxHashesWithGasProvidedSinceLastReset(key []byte) [][]byte {
	gc.mutGasProvided.RLock()
	defer gc.mutGasProvided.RUnlock()
	return gc.txHashesWithGasProvidedSinceLastReset[string(key)]
}

func (gc *gasComputation) getTxHashesWithGasProvidedAsScheduledSinceLastReset(key []byte) [][]byte {
	gc.mutGasProvided.RLock()
	defer gc.mutGasProvided.RUnlock()
	return gc.txHashesWithGasProvidedAsScheduledSinceLastReset[string(key)]
}

func (gc *gasComputation) getTxHashesWithGasRefundedSinceLastReset(key []byte) [][]byte {
	gc.mutGasRefunded.RLock()
	defer gc.mutGasRefunded.RUnlock()
	return gc.txHashesWithGasRefundedSinceLastReset[string(key)]
}

func (gc *gasComputation) getTxHashesWithGasPenalizedSinceLastReset(key []byte) [][]byte {
	gc.mutGasPenalized.RLock()
	defer gc.mutGasPenalized.RUnlock()
	return gc.txHashesWithGasPenalizedSinceLastReset[string(key)]
}

// ComputeGasProvidedByMiniBlock computes gas provided by the given miniblock in sender and receiver shard
func (gc *gasComputation) ComputeGasProvidedByMiniBlock(
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
) (uint64, uint64, error) {

	gasProvidedByMiniBlockInSenderShard := uint64(0)
	gasProvidedByMiniBlockInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		txHandler, ok := mapHashTx[string(txHash)]
		if !ok {
			log.Debug("missing transaction in ComputeGasProvidedByMiniBlock ", "type", miniBlock.Type, "txHash", txHash)
			return 0, 0, process.ErrMissingTransaction
		}

		gasProvidedByTxInSenderShard, gasProvidedByTxInReceiverShard, err := gc.ComputeGasProvidedByTx(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHandler)
		if err != nil {
			return 0, 0, err
		}

		gasProvidedByMiniBlockInSenderShard += gasProvidedByTxInSenderShard
		gasProvidedByMiniBlockInReceiverShard += gasProvidedByTxInReceiverShard
	}

	return gasProvidedByMiniBlockInSenderShard, gasProvidedByMiniBlockInReceiverShard, nil
}

// ComputeGasProvidedByTx computes gas provided by the given transaction in sender and receiver shard
func (gc *gasComputation) ComputeGasProvidedByTx(
	txSenderShardId uint32,
	txReceiverShardId uint32,
	txHandler data.TransactionHandler,
) (uint64, uint64, error) {

	if check.IfNil(txHandler) {
		return 0, 0, process.ErrNilTransaction
	}

	isGasComputeV2FlagEnabled := gc.enableEpochsHandler.IsSCDeployFlagEnabled()
	if !isGasComputeV2FlagEnabled {
		return gc.computeGasProvidedByTxV1(txSenderShardId, txReceiverShardId, txHandler)
	}

	if txHandler.GetGasLimit() == 0 {
		return 0, 0, nil
	}

	moveBalanceConsumption := gc.economicsFee.ComputeGasLimit(txHandler)
	if moveBalanceConsumption > txHandler.GetGasLimit() {
		log.Trace("ComputeGasProvidedByTx less gasLimit than moveBalance", "gasLimit", txHandler.GetGasLimit(), "moveBalanceGasCost", moveBalanceConsumption)
		return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
	}

	txTypeSndShard, txTypeDstShard := gc.txTypeHandler.ComputeTransactionType(txHandler)
	isSCCall := txTypeDstShard == process.SCDeployment ||
		txTypeDstShard == process.SCInvoking ||
		txTypeDstShard == process.BuiltInFunctionCall
	if isSCCall {
		isCrossShardSCCall := txSenderShardId != txReceiverShardId &&
			txTypeSndShard == process.MoveBalance &&
			txTypeDstShard == process.SCInvoking
		if isCrossShardSCCall {
			return moveBalanceConsumption, txHandler.GetGasLimit(), nil
		}

		return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
	}

	if gc.isRelayedTx(txTypeSndShard) {
		return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
	}

	return moveBalanceConsumption, moveBalanceConsumption, nil
}

func (gc *gasComputation) computeGasProvidedByTxV1(
	txSenderShardId uint32,
	txReceiverShardId uint32,
	txHandler data.TransactionHandler,
) (uint64, uint64, error) {
	moveBalanceConsumption := gc.economicsFee.ComputeGasLimit(txHandler)

	txTypeInShard, _ := gc.txTypeHandler.ComputeTransactionType(txHandler)
	isSCCall := txTypeInShard == process.SCDeployment ||
		txTypeInShard == process.SCInvoking ||
		txTypeInShard == process.BuiltInFunctionCall ||
		(core.IsSmartContractAddress(txHandler.GetRcvAddr()) && len(txHandler.GetData()) > 0)
	if isSCCall {
		isCrossShardSCCall := txSenderShardId != txReceiverShardId &&
			moveBalanceConsumption < txHandler.GetGasLimit() &&
			txTypeInShard != process.BuiltInFunctionCall
		if isCrossShardSCCall {
			gasProvidedByTxInSenderShard := moveBalanceConsumption
			gasProvidedByTxInReceiverShard := txHandler.GetGasLimit() - moveBalanceConsumption

			return gasProvidedByTxInSenderShard, gasProvidedByTxInReceiverShard, nil
		}

		return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
	}

	return moveBalanceConsumption, moveBalanceConsumption, nil
}

func (gc *gasComputation) isRelayedTx(txType process.TransactionType) bool {
	return txType == process.RelayedTx || txType == process.RelayedTxV2
}

// IsInterfaceNil returns true if there is no value under the interface
func (gc *gasComputation) IsInterfaceNil() bool {
	return gc == nil
}
