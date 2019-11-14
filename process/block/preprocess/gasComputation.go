package preprocess

import (
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
	}, nil
}

func (gc *gasComputation) InitGasConsumed() {
	gc.gasConsumed = 0
}

func (gc *gasComputation) AddGasConsumed(gasConsumed uint64) {
	gc.gasConsumed += gasConsumed
}

func (gc *gasComputation) SetGasConsumed(gasConsumed uint64) {
	gc.gasConsumed = gasConsumed
}

func (gc *gasComputation) GasConsumed() uint64 {
	return gc.gasConsumed
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

	if core.IsSmartContractAddress(tx.RcvAddr) {
		gasConsumedByTxInSenderShard := gc.economicsFee.ComputeGasLimit(tx)
		gasConsumedByTxInReceiverShard := tx.GasLimit

		if txSenderShardId != txReceiverShardId {
			return gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, nil
		}

		gasConsumedByTx := gasConsumedByTxInSenderShard + gasConsumedByTxInReceiverShard
		return gasConsumedByTx, gasConsumedByTx, nil
	}

	gasConsumedByTx := gc.economicsFee.ComputeGasLimit(tx)
	return gasConsumedByTx, gasConsumedByTx, nil
}

func (gc *gasComputation) IsInterfaceNil() bool {
	return gc == nil
}
