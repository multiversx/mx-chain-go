package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/core"
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

func (gc *gasComputation) GetGasConsumed() uint64 {
	return gc.gasConsumed
}

func (gc *gasComputation) ComputeGasConsumedByMiniBlockInShard(
	shardId uint32,
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
) (uint64, error) {

	gasConsumedByMiniBlockInSenderShard, gasConsumedByMiniBlockInReceiverShard, err := gc.ComputeGasConsumedByMiniBlock(
		miniBlock,
		mapHashTx)

	if err != nil {
		return 0, err
	}

	gasConsumedByMiniBlock, err := gc.ComputeGasConsumedInShard(
		shardId,
		miniBlock.SenderShardID,
		miniBlock.ReceiverShardID,
		gasConsumedByMiniBlockInSenderShard,
		gasConsumedByMiniBlockInReceiverShard)

	return gasConsumedByMiniBlock, err
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

func (gc *gasComputation) ComputeGasConsumedByTxInShard(
	shardId uint32,
	txSenderShardId uint32,
	txReceiverShardId uint32,
	txHandler data.TransactionHandler,
) (uint64, error) {

	gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, err := gc.ComputeGasConsumedByTx(
		txSenderShardId,
		txReceiverShardId,
		txHandler)
	if err != nil {
		return 0, err
	}

	gasConsumedByTx, err := gc.ComputeGasConsumedInShard(
		shardId,
		txSenderShardId,
		txReceiverShardId,
		gasConsumedByTxInSenderShard,
		gasConsumedByTxInReceiverShard)

	return gasConsumedByTx, nil
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

func (gc *gasComputation) ComputeGasConsumedInShard(
	shardId uint32,
	senderShardId uint32,
	receiverShardId uint32,
	gasConsumedInSenderShard uint64,
	gasConsumedInReceiverShard uint64,
) (uint64, error) {

	if shardId == senderShardId {
		return gasConsumedInSenderShard, nil
	}
	if shardId == receiverShardId {
		return gasConsumedInReceiverShard, nil
	}

	return 0, process.ErrInvalidShardId
}

func (gc *gasComputation) IsMaxGasLimitReached(
	gasConsumedByTxInSenderShard uint64,
	gasConsumedByTxInReceiverShard uint64,
	gasConsumedByTxInSelfShard uint64,
	currentGasConsumedByMiniBlockInSenderShard uint64,
	currentGasConsumedByMiniBlockInReceiverShard uint64,
	currentGasConsumedByBlockInSelfShard uint64,
	maxGasLimitPerBlock uint64,
) bool {

	gasConsumedByMiniBlockInSenderShard := currentGasConsumedByMiniBlockInSenderShard + gasConsumedByTxInSenderShard
	gasConsumedByMiniBlockInReceiverShard := currentGasConsumedByMiniBlockInReceiverShard + gasConsumedByTxInReceiverShard
	gasConsumedByBlockInSelfShard := currentGasConsumedByBlockInSelfShard + gasConsumedByTxInSelfShard

	isGasLimitPerMiniBlockInSenderShardReached := gasConsumedByMiniBlockInSenderShard > maxGasLimitPerBlock
	isGasLimitPerMiniBlockInReceiverShardReached := gasConsumedByMiniBlockInReceiverShard > maxGasLimitPerBlock
	isGasLimitPerBlockInSelfShardReached := gasConsumedByBlockInSelfShard > maxGasLimitPerBlock

	isMaxGasLimitReached := isGasLimitPerMiniBlockInSenderShardReached ||
		isGasLimitPerMiniBlockInReceiverShardReached ||
		isGasLimitPerBlockInSelfShardReached

	return isMaxGasLimitReached
}

func (gc *gasComputation) IsInterfaceNil() bool {
	return gc == nil
}
