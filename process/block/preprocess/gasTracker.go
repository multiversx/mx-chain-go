package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

type gasTracker struct {
	shardCoordinator sharding.Coordinator
	economicsFee     process.FeeHandler
	gasHandler       process.GasHandler
}

func (gt *gasTracker) computeGasProvided(
	senderShardId uint32,
	receiverShardId uint32,
	tx data.TransactionHandler,
	txHash []byte,
	gasInfo *gasConsumedInfo,
) (uint64, error) {
	gasProvidedByTxInSenderShard, gasProvidedByTxInReceiverShard, err := gt.computeGasProvidedByTx(
		senderShardId,
		receiverShardId,
		tx,
		txHash)
	if err != nil {
		return 0, err
	}

	gasProvidedByTxInSelfShard := uint64(0)
	if gt.shardCoordinator.SelfId() == senderShardId {
		gasProvidedByTxInSelfShard = gasProvidedByTxInSenderShard

		if gasProvidedByTxInReceiverShard > gt.economicsFee.MaxGasLimitPerTx() {
			return 0, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached
		}

		if gasInfo.gasConsumedByMiniBlockInReceiverShard+gasProvidedByTxInReceiverShard > gt.economicsFee.MaxGasLimitPerBlockForSafeCrossShard() {
			return 0, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
		}
	} else {
		gasProvidedByTxInSelfShard = gasProvidedByTxInReceiverShard
	}

	if gasInfo.totalGasConsumedInSelfShard+gasProvidedByTxInSelfShard > gt.economicsFee.MaxGasLimitPerBlock(gt.shardCoordinator.SelfId()) {
		return 0, process.ErrMaxGasLimitPerBlockInSelfShardIsReached
	}

	gasInfo.gasConsumedByMiniBlocksInSenderShard += gasProvidedByTxInSenderShard
	gasInfo.gasConsumedByMiniBlockInReceiverShard += gasProvidedByTxInReceiverShard
	gasInfo.totalGasConsumedInSelfShard += gasProvidedByTxInSelfShard

	return gasProvidedByTxInSelfShard, nil
}

func (gt *gasTracker) computeGasProvidedByTx(
	senderShardId uint32,
	receiverShardId uint32,
	tx data.TransactionHandler,
	txHash []byte,
) (uint64, uint64, error) {

	txGasLimitInSenderShard, txGasLimitInReceiverShard, err := gt.gasHandler.ComputeGasProvidedByTx(
		senderShardId,
		receiverShardId,
		tx)
	if err != nil {
		return 0, 0, err
	}

	if core.IsSmartContractAddress(tx.GetRcvAddr()) {
		txGasRefunded := gt.gasHandler.GasRefunded(txHash)
		txGasPenalized := gt.gasHandler.GasPenalized(txHash)
		txGasToBeSubtracted := txGasRefunded + txGasPenalized
		if txGasLimitInReceiverShard < txGasToBeSubtracted {
			return 0, 0, process.ErrInsufficientGasLimitInTx
		}

		if senderShardId == receiverShardId {
			txGasLimitInSenderShard -= txGasToBeSubtracted
			txGasLimitInReceiverShard -= txGasToBeSubtracted
		}
	}

	return txGasLimitInSenderShard, txGasLimitInReceiverShard, nil
}
