package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const noOverestimationFactor = uint64(100)

type gasTracker struct {
	shardCoordinator    sharding.Coordinator
	economicsFee        process.FeeHandler
	gasHandler          process.GasHandler
	enableEpochsHandler common.EnableEpochsHandler
	enableRoundsHandler common.EnableRoundsHandler
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

		if gasProvidedByTxInReceiverShard > gt.getMaxGasLimitPerTx() {
			return 0, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached
		}

		if gasInfo.gasConsumedByMiniBlockInReceiverShard+gasProvidedByTxInReceiverShard > gt.getMaxGasLimitPerBlockForSafeCrossShard() {
			return 0, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
		}
	} else {
		gasProvidedByTxInSelfShard = gasProvidedByTxInReceiverShard
	}

	if gasInfo.totalGasConsumedInSelfShard+gasProvidedByTxInSelfShard > gt.getMaxGasLimitPerBlock() {
		return 0, process.ErrMaxGasLimitPerBlockInSelfShardIsReached
	}

	gasInfo.gasConsumedByMiniBlocksInSenderShard += gasProvidedByTxInSenderShard
	gasInfo.gasConsumedByMiniBlockInReceiverShard += gasProvidedByTxInReceiverShard
	gasInfo.totalGasConsumedInSelfShard += gasProvidedByTxInSelfShard

	return gasProvidedByTxInSelfShard, nil
}

func (gt *gasTracker) getEpochAndOverestimationFactorForGasLimits() (epoch uint32, overestimationFactor uint64) {
	epoch = gt.enableEpochsHandler.GetCurrentEpoch()
	overestimationFactor = noOverestimationFactor
	isSupernovaRoundEnabled := gt.enableRoundsHandler.IsFlagEnabled(common.SupernovaRoundFlag)
	if !isSupernovaRoundEnabled {
		return
	}

	// new limits and overestimation should be enabled once the Supernova round is active
	epoch = epoch - 1
	overestimationFactor = gt.economicsFee.BlockCapacityOverestimationFactor()

	return
}

func (gt *gasTracker) getMaxGasLimitPerTx() uint64 {
	epoch, _ := gt.getEpochAndOverestimationFactorForGasLimits()
	return gt.economicsFee.MaxGasLimitPerTxInEpoch(epoch)
}

func (gt *gasTracker) getMaxGasLimitPerBlockForSafeCrossShard() uint64 {
	epoch, overEstimationFactor := gt.getEpochAndOverestimationFactorForGasLimits()
	return gt.economicsFee.MaxGasLimitPerBlockForSafeCrossShardInEpoch(epoch) * overEstimationFactor / 100
}

func (gt *gasTracker) getMaxGasLimitPerBlock() uint64 {
	epoch, overEstimationFactor := gt.getEpochAndOverestimationFactorForGasLimits()
	return gt.economicsFee.MaxGasLimitPerBlockInEpoch(gt.shardCoordinator.SelfId(), epoch) * overEstimationFactor / 100
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
