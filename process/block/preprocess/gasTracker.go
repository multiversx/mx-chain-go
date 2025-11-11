package preprocess

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const noOverestimationFactor = uint64(100)

type gasTracker struct {
	shardCoordinator     sharding.Coordinator
	economicsFee         process.FeeHandler
	gasHandler           process.GasHandler
	enableEpochsHandler  common.EnableEpochsHandler
	enableRoundsHandler  common.EnableRoundsHandler
	mut                  sync.RWMutex
	overEstimationFactor uint64
	epochForLimits       uint32
	roundForLimits       uint64
}

func newGasTracker(
	shardCoordinator sharding.Coordinator,
	gasHandler process.GasHandler,
	economicsFee process.FeeHandler,
	enableEpochsHandler common.EnableEpochsHandler,
	enableRoundsHandler common.EnableRoundsHandler,
) gasTracker {
	return gasTracker{
		shardCoordinator:     shardCoordinator,
		economicsFee:         economicsFee,
		gasHandler:           gasHandler,
		enableEpochsHandler:  enableEpochsHandler,
		enableRoundsHandler:  enableRoundsHandler,
		overEstimationFactor: noOverestimationFactor,
	}
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (gt *gasTracker) EpochConfirmed(epoch uint32, _ uint64) {
	gt.mut.Lock()
	defer gt.mut.Unlock()

	gt.epochForLimits = epoch

	// if already computed, only store the new epoch
	if gt.overEstimationFactor != noOverestimationFactor {
		return
	}

	isEpochFlagEnabled := gt.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, gt.epochForLimits)
	if !isEpochFlagEnabled {
		return
	}

	isRoundFlagEnabled := gt.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, gt.roundForLimits)
	if !isRoundFlagEnabled && epoch > 0 {
		gt.epochForLimits = epoch - 1 // use the previous epoch until activation round
	}
}

// RoundConfirmed is called whenever a new round is confirmed
func (gt *gasTracker) RoundConfirmed(round uint64, _ uint64) {
	gt.mut.Lock()
	defer gt.mut.Unlock()

	gt.roundForLimits = round

	// if already computed, only store the new round
	if gt.overEstimationFactor != noOverestimationFactor {
		return
	}

	isRoundFlagEnabled := gt.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, gt.roundForLimits)
	if !isRoundFlagEnabled {
		return
	}

	// new limits and overestimation should be enabled once the Supernova round is active
	gt.overEstimationFactor = gt.economicsFee.BlockCapacityOverestimationFactor()
	gt.epochForLimits = gt.epochForLimits + 1 // we save prev epoch until round is enabled too
}

// IsInterfaceNil returns true if there is no value under the interface
func (gt *gasTracker) IsInterfaceNil() bool {
	return gt == nil
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

	gt.mut.RLock()
	epoch := gt.epochForLimits
	overEstimationFactor := gt.overEstimationFactor
	gt.mut.RUnlock()

	gasProvidedByTxInSelfShard := uint64(0)
	if gt.shardCoordinator.SelfId() == senderShardId {
		gasProvidedByTxInSelfShard = gasProvidedByTxInSenderShard

		if gasProvidedByTxInReceiverShard > gt.getMaxGasLimitPerTx(epoch) {
			return 0, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached
		}

		if gasInfo.gasConsumedByMiniBlockInReceiverShard+gasProvidedByTxInReceiverShard > gt.getMaxGasLimitPerBlockForSafeCrossShard(epoch, overEstimationFactor) {
			return 0, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
		}
	} else {
		gasProvidedByTxInSelfShard = gasProvidedByTxInReceiverShard
	}

	if gasInfo.totalGasConsumedInSelfShard+gasProvidedByTxInSelfShard > gt.getMaxGasLimitPerBlock(epoch, overEstimationFactor) {
		return 0, process.ErrMaxGasLimitPerBlockInSelfShardIsReached
	}

	gasInfo.gasConsumedByMiniBlocksInSenderShard += gasProvidedByTxInSenderShard
	gasInfo.gasConsumedByMiniBlockInReceiverShard += gasProvidedByTxInReceiverShard
	gasInfo.totalGasConsumedInSelfShard += gasProvidedByTxInSelfShard

	return gasProvidedByTxInSelfShard, nil
}

func (gt *gasTracker) getMaxGasLimitPerTx(epoch uint32) uint64 {
	return gt.economicsFee.MaxGasLimitPerTxInEpoch(epoch)
}

func (gt *gasTracker) getMaxGasLimitPerBlockForSafeCrossShard(epoch uint32, overEstimationFactor uint64) uint64 {
	return gt.economicsFee.MaxGasLimitPerBlockForSafeCrossShardInEpoch(epoch) * overEstimationFactor / 100
}

func (gt *gasTracker) getMaxGasLimitPerBlock(epoch uint32, overEstimationFactor uint64) uint64 {
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
