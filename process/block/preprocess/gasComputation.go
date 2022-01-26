package preprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.GasHandler = (*gasComputation)(nil)

type gasComputation struct {
	economicsFee  process.FeeHandler
	txTypeHandler process.TxTypeHandler
	//TODO: Refactor these mutexes and maps in separated structures that handle the locking and unlocking for each operation required
	gasProvided                                      map[string]uint64
	txHashesWithGasProvidedSinceLastReset            [][]byte
	gasProvidedAsScheduled                           map[string]uint64
	txHashesWithGasProvidedAsScheduledSinceLastReset [][]byte
	mutGasProvided                                   sync.RWMutex
	gasRefunded                                      map[string]uint64
	txHashesWithGasRefundedSinceLastReset            [][]byte
	mutGasRefunded                                   sync.RWMutex
	gasPenalized                                     map[string]uint64
	txHashesWithGasPenalizedSinceLastReset           [][]byte
	mutGasPenalized                                  sync.RWMutex

	flagGasComputeV2        atomic.Flag
	gasComputeV2EnableEpoch uint32
}

// NewGasComputation creates a new object which computes the gas consumption
func NewGasComputation(
	economicsFee process.FeeHandler,
	txTypeHandler process.TxTypeHandler,
	epochNotifier process.EpochNotifier,
	gasComputeV2EnableEpoch uint32,
) (*gasComputation, error) {
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	g := &gasComputation{
		txTypeHandler:                         txTypeHandler,
		economicsFee:                          economicsFee,
		gasProvided:                           make(map[string]uint64),
		txHashesWithGasProvidedSinceLastReset: make([][]byte, 0),
		gasProvidedAsScheduled:                make(map[string]uint64),
		txHashesWithGasProvidedAsScheduledSinceLastReset: make([][]byte, 0),
		gasRefunded:                            make(map[string]uint64),
		txHashesWithGasRefundedSinceLastReset:  make([][]byte, 0),
		gasPenalized:                           make(map[string]uint64),
		txHashesWithGasPenalizedSinceLastReset: make([][]byte, 0),
		gasComputeV2EnableEpoch:                gasComputeV2EnableEpoch,
	}
	log.Debug("gasComputation: enable epoch for sc deploy", "epoch", g.gasComputeV2EnableEpoch)

	epochNotifier.RegisterNotifyHandler(g)

	return g, nil
}

// Init method inits provided, refunded and penalized gas structures
func (gc *gasComputation) Init() {
	gc.mutGasProvided.Lock()
	gc.gasProvided = make(map[string]uint64)
	gc.txHashesWithGasProvidedSinceLastReset = make([][]byte, 0)
	gc.gasProvidedAsScheduled = make(map[string]uint64)
	gc.txHashesWithGasProvidedAsScheduledSinceLastReset = make([][]byte, 0)
	gc.mutGasProvided.Unlock()

	gc.mutGasRefunded.Lock()
	gc.gasRefunded = make(map[string]uint64)
	gc.txHashesWithGasRefundedSinceLastReset = make([][]byte, 0)
	gc.mutGasRefunded.Unlock()

	gc.mutGasPenalized.Lock()
	gc.gasPenalized = make(map[string]uint64)
	gc.txHashesWithGasPenalizedSinceLastReset = make([][]byte, 0)
	gc.mutGasPenalized.Unlock()
}

// Reset method resets tx hashes with gas provided, refunded and penalized since last reset
func (gc *gasComputation) Reset() {
	gc.mutGasProvided.Lock()
	gc.txHashesWithGasProvidedSinceLastReset = make([][]byte, 0)
	gc.txHashesWithGasProvidedAsScheduledSinceLastReset = make([][]byte, 0)
	gc.mutGasProvided.Unlock()

	gc.mutGasRefunded.Lock()
	gc.txHashesWithGasRefundedSinceLastReset = make([][]byte, 0)
	gc.mutGasRefunded.Unlock()

	gc.mutGasPenalized.Lock()
	gc.txHashesWithGasPenalizedSinceLastReset = make([][]byte, 0)
	gc.mutGasPenalized.Unlock()
}

// SetGasProvided sets gas provided for a given hash
func (gc *gasComputation) SetGasProvided(gasProvided uint64, hash []byte) {
	gc.mutGasProvided.Lock()
	gc.gasProvided[string(hash)] = gasProvided
	gc.txHashesWithGasProvidedSinceLastReset = append(gc.txHashesWithGasProvidedSinceLastReset, hash)
	gc.mutGasProvided.Unlock()
}

// SetGasProvidedAsScheduled sets gas provided as scheduled for a given hash
func (gc *gasComputation) SetGasProvidedAsScheduled(gasProvided uint64, hash []byte) {
	gc.mutGasProvided.Lock()
	gc.gasProvidedAsScheduled[string(hash)] = gasProvided
	gc.txHashesWithGasProvidedAsScheduledSinceLastReset = append(gc.txHashesWithGasProvidedAsScheduledSinceLastReset, hash)
	gc.mutGasProvided.Unlock()
}

// SetGasRefunded sets gas refunded for a given hash
func (gc *gasComputation) SetGasRefunded(gasRefunded uint64, hash []byte) {
	gc.mutGasRefunded.Lock()
	gc.gasRefunded[string(hash)] = gasRefunded
	gc.txHashesWithGasRefundedSinceLastReset = append(gc.txHashesWithGasRefundedSinceLastReset, hash)
	gc.mutGasRefunded.Unlock()

	//TODO: Remove it or set log level to Trace
	log.Debug("gasComputation.SetGasRefunded", "tx hash", hash, "gas refunded", gasRefunded)
}

// SetGasPenalized sets gas penalized for a given hash
func (gc *gasComputation) SetGasPenalized(gasPenalized uint64, hash []byte) {
	gc.mutGasPenalized.Lock()
	gc.gasPenalized[string(hash)] = gasPenalized
	gc.txHashesWithGasPenalizedSinceLastReset = append(gc.txHashesWithGasPenalizedSinceLastReset, hash)
	gc.mutGasPenalized.Unlock()

	//TODO: Remove it or set log level to Trace
	log.Debug("gasComputation.SetGasPenalized", "tx hash", hash, "gas penalized", gasPenalized)
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
func (gc *gasComputation) RestoreGasSinceLastReset() {
	gc.RemoveGasProvided(gc.getTxHashesWithGasProvidedSinceLastReset())
	gc.RemoveGasProvidedAsScheduled(gc.getTxHashesWithGasProvidedAsScheduledSinceLastReset())
	gc.RemoveGasRefunded(gc.getTxHashesWithGasRefundedSinceLastReset())
	gc.RemoveGasPenalized(gc.getTxHashesWithGasPenalizedSinceLastReset())
}

func (gc *gasComputation) getTxHashesWithGasProvidedSinceLastReset() [][]byte {
	gc.mutGasProvided.RLock()
	defer gc.mutGasProvided.RUnlock()
	return gc.txHashesWithGasProvidedSinceLastReset
}

func (gc *gasComputation) getTxHashesWithGasProvidedAsScheduledSinceLastReset() [][]byte {
	gc.mutGasProvided.RLock()
	defer gc.mutGasProvided.RUnlock()
	return gc.txHashesWithGasProvidedAsScheduledSinceLastReset
}

func (gc *gasComputation) getTxHashesWithGasRefundedSinceLastReset() [][]byte {
	gc.mutGasRefunded.RLock()
	defer gc.mutGasRefunded.RUnlock()
	return gc.txHashesWithGasRefundedSinceLastReset
}

func (gc *gasComputation) getTxHashesWithGasPenalizedSinceLastReset() [][]byte {
	gc.mutGasPenalized.RLock()
	defer gc.mutGasPenalized.RUnlock()
	return gc.txHashesWithGasPenalizedSinceLastReset
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

	if !gc.flagGasComputeV2.IsSet() {
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

// EpochConfirmed is called whenever a new epoch is confirmed
func (gc *gasComputation) EpochConfirmed(epoch uint32, _ uint64) {
	gc.flagGasComputeV2.SetValue(epoch >= gc.gasComputeV2EnableEpoch)
	log.Debug("gasComputation: compute v2", "enabled", gc.flagGasComputeV2.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (gc *gasComputation) IsInterfaceNil() bool {
	return gc == nil
}
