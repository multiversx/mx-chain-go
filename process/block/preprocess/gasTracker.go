package preprocess

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type gasTracker struct {
	shardCoordinator           sharding.Coordinator
	economicsFee               process.FeeHandler
	gasHandler                 process.GasHandler
}

func (gt *gasTracker) computeGasConsumed(
	senderShardId uint32,
	receiverShardId uint32,
	tx data.TransactionHandler,
	txHash []byte,
	gasInfo *gasConsumedInfo,
) error {
	gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, err := gt.computeGasConsumedByTx(
		senderShardId,
		receiverShardId,
		tx,
		txHash)
	if err != nil {
		return err
	}

	gasConsumedByTxInSelfShard := uint64(0)
	if gt.shardCoordinator.SelfId() == senderShardId {
		gasConsumedByTxInSelfShard = gasConsumedByTxInSenderShard

		if gasInfo.gasConsumedByMiniBlockInReceiverShard+gasConsumedByTxInReceiverShard > gt.economicsFee.MaxGasLimitPerBlock(gt.shardCoordinator.SelfId()) {
			return process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
		}
	} else {
		gasConsumedByTxInSelfShard = gasConsumedByTxInReceiverShard

		if gasInfo.gasConsumedByMiniBlocksInSenderShard+gasConsumedByTxInSenderShard > gt.economicsFee.MaxGasLimitPerBlock(senderShardId) {
			return process.ErrMaxGasLimitPerMiniBlockInSenderShardIsReached
		}
	}

	if gasInfo.totalGasConsumedInSelfShard+gasConsumedByTxInSelfShard > gt.economicsFee.MaxGasLimitPerBlock(gt.shardCoordinator.SelfId()) {
		return process.ErrMaxGasLimitPerBlockInSelfShardIsReached
	}

	gasInfo.gasConsumedByMiniBlocksInSenderShard += gasConsumedByTxInSenderShard
	gasInfo.gasConsumedByMiniBlockInReceiverShard += gasConsumedByTxInReceiverShard
	gasInfo.totalGasConsumedInSelfShard += gasConsumedByTxInSelfShard
	gt.gasHandler.SetGasConsumed(gasConsumedByTxInSelfShard, txHash)

	return nil
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
