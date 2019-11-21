package preprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type gasComputation struct {
	economicsFee   process.FeeHandler
	gasConsumed    map[string]uint64
	mutGasConsumed sync.RWMutex
	gasRefunded    map[string]uint64
	mutGasRefunded sync.RWMutex
}

func NewGasComputation(
	economicsFee process.FeeHandler,
) (*gasComputation, error) {

	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	return &gasComputation{
		economicsFee: economicsFee,
		gasConsumed:  make(map[string]uint64),
		gasRefunded:  make(map[string]uint64),
	}, nil
}

func (gc *gasComputation) Init() {
	gc.mutGasConsumed.Lock()
	gc.gasConsumed = make(map[string]uint64)
	gc.mutGasConsumed.Unlock()

	gc.mutGasRefunded.Lock()
	gc.gasRefunded = make(map[string]uint64)
	gc.mutGasRefunded.Unlock()
}

func (gc *gasComputation) SetGasConsumed(gasConsumed uint64, hash []byte) {
	gc.mutGasConsumed.Lock()
	gc.gasConsumed[string(hash)] = gasConsumed
	gc.mutGasConsumed.Unlock()
}

func (gc *gasComputation) SetGasRefunded(gasRefunded uint64, hash []byte) {
	gc.mutGasRefunded.Lock()
	gc.gasRefunded[string(hash)] = gasRefunded
	gc.mutGasRefunded.Unlock()
}

func (gc *gasComputation) GasConsumed(hash []byte) uint64 {
	gc.mutGasConsumed.RLock()
	defer gc.mutGasConsumed.RUnlock()

	return gc.gasConsumed[string(hash)]

}

func (gc *gasComputation) GasRefunded(hash []byte) uint64 {
	gc.mutGasRefunded.RLock()
	defer gc.mutGasRefunded.RUnlock()

	return gc.gasRefunded[string(hash)]
}

func (gc *gasComputation) TotalGasConsumed() uint64 {
	gc.mutGasConsumed.RLock()
	defer gc.mutGasConsumed.RUnlock()

	totalGasConsumed := uint64(0)
	for _, gasConsumed := range gc.gasConsumed {
		totalGasConsumed += gasConsumed
	}

	return totalGasConsumed
}

func (gc *gasComputation) TotalGasRefunded() uint64 {
	gc.mutGasRefunded.RLock()
	defer gc.mutGasRefunded.RUnlock()

	totalGasRefunded := uint64(0)
	for _, gasRefunded := range gc.gasRefunded {
		totalGasRefunded += gasRefunded
	}

	return totalGasRefunded
}

func (gc *gasComputation) RemoveGasConsumed(hashes [][]byte) {
	gc.mutGasConsumed.Lock()
	for _, hash := range hashes {
		delete(gc.gasConsumed, string(hash))
	}
	gc.mutGasConsumed.Unlock()
}

func (gc *gasComputation) RemoveGasRefunded(hashes [][]byte) {
	gc.mutGasRefunded.Lock()
	for _, hash := range hashes {
		delete(gc.gasRefunded, string(hash))
	}
	gc.mutGasRefunded.Unlock()
}

func (gc *gasComputation) ComputeGasConsumedByMiniBlock(
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
) (uint64, uint64, error) {

	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		txHandler, ok := mapHashTx[string(txHash)]
		if !ok {
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

func (gc *gasComputation) ComputeGasConsumedByTx(
	txSenderShardId uint32,
	txReceiverShardId uint32,
	txHandler data.TransactionHandler,
) (uint64, uint64, error) {

	tx, ok := txHandler.(*transaction.Transaction)
	if !ok {
		return 0, 0, process.ErrWrongTypeAssertion
	}

	txGasLimitConsumption := gc.economicsFee.ComputeGasLimit(tx)
	if tx.GasLimit < txGasLimitConsumption {
		return 0, 0, process.ErrInsufficientGasLimitInTx
	}

	if core.IsSmartContractAddress(tx.RcvAddr) {
		if txSenderShardId != txReceiverShardId {
			gasConsumedByTxInSenderShard := txGasLimitConsumption
			gasConsumedByTxInReceiverShard := tx.GasLimit - txGasLimitConsumption

			return gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, nil
		}

		return tx.GasLimit, tx.GasLimit, nil
	}

	return txGasLimitConsumption, txGasLimitConsumption, nil
}

func (gc *gasComputation) IsInterfaceNil() bool {
	return gc == nil
}
