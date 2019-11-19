package preprocess

import (
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type gasComputation struct {
	economicsFee process.FeeHandler
	gasConsumed  uint64
	gasRefunded  uint64
}

func NewGasComputation(
	economicsFee process.FeeHandler,
) (*gasComputation, error) {

	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	return &gasComputation{
		economicsFee: economicsFee,
		gasConsumed:  0,
		gasRefunded:  0,
	}, nil
}

func (gc *gasComputation) Init() {
	atomic.StoreUint64(&gc.gasConsumed, 0)
	atomic.StoreUint64(&gc.gasRefunded, 0)
}

func (gc *gasComputation) AddGasConsumed(gasConsumed uint64) {
	atomic.AddUint64(&gc.gasConsumed, gasConsumed)
}

func (gc *gasComputation) AddGasRefunded(gasRefunded uint64) {
	atomic.AddUint64(&gc.gasRefunded, gasRefunded)
}

func (gc *gasComputation) SetGasConsumed(gasConsumed uint64) {
	atomic.StoreUint64(&gc.gasConsumed, gasConsumed)
}

func (gc *gasComputation) SetGasRefunded(gasRefunded uint64) {
	atomic.StoreUint64(&gc.gasRefunded, gasRefunded)
}

func (gc *gasComputation) GasConsumed() uint64 {
	return atomic.LoadUint64(&gc.gasConsumed)
}

func (gc *gasComputation) GasRefunded() uint64 {
	return atomic.LoadUint64(&gc.gasRefunded)
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
