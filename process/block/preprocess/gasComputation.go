package preprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.GasHandler = (*gasComputation)(nil)

type gasComputation struct {
	economicsFee   process.FeeHandler
	txTypeHandler  process.TxTypeHandler
	gasConsumed    map[string]uint64
	mutGasConsumed sync.RWMutex
	gasRefunded    map[string]uint64
	mutGasRefunded sync.RWMutex

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
		txTypeHandler:           txTypeHandler,
		economicsFee:            economicsFee,
		gasConsumed:             make(map[string]uint64),
		gasRefunded:             make(map[string]uint64),
		gasComputeV2EnableEpoch: gasComputeV2EnableEpoch,
	}

	epochNotifier.RegisterNotifyHandler(g)

	return g, nil
}

// Init method resets consumed and refunded gas structures
func (gc *gasComputation) Init() {
	gc.mutGasConsumed.Lock()
	gc.gasConsumed = make(map[string]uint64)
	gc.mutGasConsumed.Unlock()

	gc.mutGasRefunded.Lock()
	gc.gasRefunded = make(map[string]uint64)
	gc.mutGasRefunded.Unlock()
}

// SetGasConsumed sets gas consumed for a given hash
func (gc *gasComputation) SetGasConsumed(gasConsumed uint64, hash []byte) {
	gc.mutGasConsumed.Lock()
	gc.gasConsumed[string(hash)] = gasConsumed
	gc.mutGasConsumed.Unlock()
}

// SetGasRefunded sets gas refunded for a given hash
func (gc *gasComputation) SetGasRefunded(gasRefunded uint64, hash []byte) {
	gc.mutGasRefunded.Lock()
	gc.gasRefunded[string(hash)] = gasRefunded
	gc.mutGasRefunded.Unlock()
}

// GasConsumed gets gas consumed for a given hash
func (gc *gasComputation) GasConsumed(hash []byte) uint64 {
	gc.mutGasConsumed.RLock()
	gasConsumed := gc.gasConsumed[string(hash)]
	gc.mutGasConsumed.RUnlock()

	return gasConsumed

}

// GasRefunded gets gas refunded for a given hash
func (gc *gasComputation) GasRefunded(hash []byte) uint64 {
	gc.mutGasRefunded.RLock()
	gasRefunded := gc.gasRefunded[string(hash)]
	gc.mutGasRefunded.RUnlock()

	return gasRefunded
}

// TotalGasConsumed gets the total gas consumed
func (gc *gasComputation) TotalGasConsumed() uint64 {
	totalGasConsumed := uint64(0)

	gc.mutGasConsumed.RLock()
	for _, gasConsumed := range gc.gasConsumed {
		totalGasConsumed += gasConsumed
	}
	gc.mutGasConsumed.RUnlock()

	return totalGasConsumed
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

// RemoveGasConsumed removes gas consumed for the given hashes
func (gc *gasComputation) RemoveGasConsumed(hashes [][]byte) {
	gc.mutGasConsumed.Lock()
	for _, hash := range hashes {
		delete(gc.gasConsumed, string(hash))
	}
	gc.mutGasConsumed.Unlock()
}

// RemoveGasRefunded removes gas refunded for the given hashes
func (gc *gasComputation) RemoveGasRefunded(hashes [][]byte) {
	gc.mutGasRefunded.Lock()
	for _, hash := range hashes {
		delete(gc.gasRefunded, string(hash))
	}
	gc.mutGasRefunded.Unlock()
}

// ComputeGasConsumedByMiniBlock computes gas consumed by the given miniblock in sender and receiver shard
func (gc *gasComputation) ComputeGasConsumedByMiniBlock(
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
) (uint64, uint64, error) {

	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		txHandler, ok := mapHashTx[string(txHash)]
		if !ok {
			log.Debug("missing transaction in ComputeGasConsumedByMiniBlock ", "type", miniBlock.Type, "txHash", txHash)
			return 0, 0, process.ErrMissingTransaction
		}

		gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, err := gc.ComputeGasConsumedByTx(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHandler)
		if err != nil {
			return 0, 0, err
		}

		gasConsumedByMiniBlockInSenderShard += gasConsumedByTxInSenderShard
		gasConsumedByMiniBlockInReceiverShard += gasConsumedByTxInReceiverShard
	}

	return gasConsumedByMiniBlockInSenderShard, gasConsumedByMiniBlockInReceiverShard, nil
}

// ComputeGasConsumedByTx computes gas consumed by the given transaction in sender and receiver shard
func (gc *gasComputation) ComputeGasConsumedByTx(
	txSenderShardId uint32,
	txReceiverShardId uint32,
	txHandler data.TransactionHandler,
) (uint64, uint64, error) {

	if check.IfNil(txHandler) {
		return 0, 0, process.ErrNilTransaction
	}

	if !gc.flagGasComputeV2.IsSet() {
		return gc.computeGasConsumedByTxV1(txSenderShardId, txReceiverShardId, txHandler)
	}

	moveBalanceConsumption := gc.economicsFee.ComputeGasLimit(txHandler)

	_, txCrossShard := gc.txTypeHandler.ComputeTransactionType(txHandler)
	isSCCall := txCrossShard == process.SCDeployment ||
		txCrossShard == process.SCInvoking ||
		txCrossShard == process.BuiltInFunctionCall
	if isSCCall {
		isCrossShardSCCall := txSenderShardId != txReceiverShardId &&
			moveBalanceConsumption < txHandler.GetGasLimit() &&
			txCrossShard != process.BuiltInFunctionCall
		if isCrossShardSCCall {
			gasConsumedByTxInSenderShard := moveBalanceConsumption
			gasConsumedByTxInReceiverShard := txHandler.GetGasLimit() - moveBalanceConsumption

			return gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, nil
		}

		return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
	}

	return moveBalanceConsumption, moveBalanceConsumption, nil
}

func (gc *gasComputation) computeGasConsumedByTxV1(
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
			gasConsumedByTxInSenderShard := moveBalanceConsumption
			gasConsumedByTxInReceiverShard := txHandler.GetGasLimit() - moveBalanceConsumption

			return gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, nil
		}

		return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
	}

	return moveBalanceConsumption, moveBalanceConsumption, nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (gc *gasComputation) EpochConfirmed(epoch uint32) {
	gc.flagGasComputeV2.Toggle(epoch >= gc.gasComputeV2EnableEpoch)
	log.Debug("gasComputation: compute v2", "enabled", gc.flagGasComputeV2.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (gc *gasComputation) IsInterfaceNil() bool {
	return gc == nil
}
