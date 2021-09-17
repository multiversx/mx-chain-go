package preprocess

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type gasTracker struct {
	shardCoordinator sharding.Coordinator
	economicsFee     process.FeeHandler
	gasHandler       process.GasHandler
}

func (gt *gasTracker) computeGasConsumed(
	senderShardId uint32,
	receiverShardId uint32,
	tx data.TransactionHandler,
	txHash []byte,
	gasInfo *process.GasConsumedInfo,
) (uint64, error) {
	gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, err := gt.computeGasConsumedByTx(
		senderShardId,
		receiverShardId,
		tx,
		txHash)
	if err != nil {
		return 0, err
	}

	gasConsumedByTxInSelfShard := uint64(0)
	if gt.shardCoordinator.SelfId() == senderShardId {
		gasConsumedByTxInSelfShard = gasConsumedByTxInSenderShard

		//TODO: If here is used selfId() instead receiverShardId the maximum gas used in mini blocks to metachain will be 1.5 bil. instead 15 bil.
		if gasInfo.GasConsumedByMiniBlockInReceiverShard+gasConsumedByTxInReceiverShard > gt.economicsFee.MaxGasLimitPerBlock(gt.shardCoordinator.SelfId()) {
			return 0, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
		}
	} else {
		gasConsumedByTxInSelfShard = gasConsumedByTxInReceiverShard

		if gasInfo.GasConsumedByMiniBlocksInSenderShard+gasConsumedByTxInSenderShard > gt.economicsFee.MaxGasLimitPerBlock(senderShardId) {
			return 0, process.ErrMaxGasLimitPerMiniBlockInSenderShardIsReached
		}
	}

	if gasInfo.TotalGasConsumedInSelfShard+gasConsumedByTxInSelfShard > gt.economicsFee.MaxGasLimitPerBlock(gt.shardCoordinator.SelfId()) {
		return 0, process.ErrMaxGasLimitPerBlockInSelfShardIsReached
	}

	gasInfo.GasConsumedByMiniBlocksInSenderShard += gasConsumedByTxInSenderShard
	gasInfo.GasConsumedByMiniBlockInReceiverShard += gasConsumedByTxInReceiverShard
	gasInfo.TotalGasConsumedInSelfShard += gasConsumedByTxInSelfShard

	return gasConsumedByTxInSelfShard, nil
}

func (gt *gasTracker) computeGasConsumedByTx(
	senderShardId uint32,
	receiverShardId uint32,
	tx data.TransactionHandler,
	txHash []byte,
) (uint64, uint64, error) {

	txGasLimitInSenderShard, txGasLimitInReceiverShard, err := gt.gasHandler.ComputeGasConsumedByTx(
		senderShardId,
		receiverShardId,
		tx)
	if err != nil {
		return 0, 0, err
	}

	if core.IsSmartContractAddress(tx.GetRcvAddr()) {
		txGasRefunded := gt.gasHandler.GasRefunded(txHash)

		if txGasLimitInReceiverShard < txGasRefunded {
			return 0, 0, process.ErrInsufficientGasLimitInTx
		}

		if senderShardId == receiverShardId {
			txGasLimitInSenderShard -= txGasRefunded
			txGasLimitInReceiverShard -= txGasRefunded
		}
	}

	return txGasLimitInSenderShard, txGasLimitInReceiverShard, nil
}

func (gt *gasTracker) computeGasConsumedByCrossScrInReceiverShard(gasInfo *process.GasConsumedInfo, tx data.TransactionHandler) error {
	gasConsumedByTxInReceiverShard := tx.GetGasLimit()
	//TODO: If here is used selfId() instead receiverShardId the maximum gas used in mini blocks to metachain will be 1.5 bil. instead 15 bil.
	if gasInfo.GasConsumedByMiniBlockInReceiverShard+gasConsumedByTxInReceiverShard > gt.economicsFee.MaxGasLimitPerBlock(gt.shardCoordinator.SelfId()) {
		return process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
	}

	gasInfo.GasConsumedByMiniBlockInReceiverShard += gasConsumedByTxInReceiverShard
	return nil
}
