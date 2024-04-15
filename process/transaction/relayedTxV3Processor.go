package transaction

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

type relayedTxV3Processor struct {
	economicsFee     process.FeeHandler
	shardCoordinator sharding.Coordinator
}

// NewRelayedTxV3Processor returns a new instance of relayedTxV3Processor
func NewRelayedTxV3Processor(economicsFee process.FeeHandler, shardCoordinator sharding.Coordinator) (*relayedTxV3Processor, error) {
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &relayedTxV3Processor{
		economicsFee:     economicsFee,
		shardCoordinator: shardCoordinator,
	}, nil
}

// CheckRelayedTx checks the relayed transaction and its inner transactions
func (proc *relayedTxV3Processor) CheckRelayedTx(tx *transaction.Transaction) error {
	if tx.GetValue().Cmp(big.NewInt(0)) != 0 {
		return process.ErrRelayedTxV3ZeroVal
	}
	if !bytes.Equal(tx.RcvAddr, tx.SndAddr) {
		return process.ErrRelayedTxV3SenderDoesNotMatchReceiver
	}
	if tx.GasLimit < proc.computeRelayedTxMinGasLimit(tx) {
		return process.ErrRelayedTxV3GasLimitMismatch
	}

	innerTxs := tx.InnerTransactions
	for _, innerTx := range innerTxs {
		if len(innerTx.RelayerAddr) == 0 {
			return process.ErrRelayedTxV3EmptyRelayer
		}
		if !bytes.Equal(innerTx.RelayerAddr, tx.SndAddr) {
			return process.ErrRelayedTxV3RelayerMismatch
		}
		if tx.GasPrice != innerTx.GasPrice {
			return process.ErrRelayedV3GasPriceMismatch
		}

		senderShard := proc.shardCoordinator.ComputeId(innerTx.SndAddr)
		relayerShard := proc.shardCoordinator.ComputeId(innerTx.RelayerAddr)
		if senderShard != relayerShard {
			return process.ErrRelayedTxV3SenderShardMismatch
		}
	}

	return nil
}

// ComputeRelayedTxFees returns the both the total fee for the entire relayed tx and the relayed only fee
func (proc *relayedTxV3Processor) ComputeRelayedTxFees(tx *transaction.Transaction) (*big.Int, *big.Int) {
	relayerMoveBalanceFee := proc.economicsFee.ComputeMoveBalanceFee(tx)
	uniqueSenders := proc.GetUniqueSendersRequiredFeesMap(tx.InnerTransactions)

	relayerFee := big.NewInt(0).Mul(relayerMoveBalanceFee, big.NewInt(int64(len(uniqueSenders))))

	totalFee := big.NewInt(0)
	for _, fee := range uniqueSenders {
		totalFee.Add(totalFee, fee)
	}
	totalFee.Add(totalFee, relayerFee)

	return relayerFee, totalFee
}

// GetUniqueSendersRequiredFeesMap returns the map of unique inner transactions senders and the required fees for all transactions
func (proc *relayedTxV3Processor) GetUniqueSendersRequiredFeesMap(innerTxs []*transaction.Transaction) map[string]*big.Int {
	uniqueSendersMap := make(map[string]*big.Int)
	for _, innerTx := range innerTxs {
		senderStr := string(innerTx.SndAddr)
		_, exists := uniqueSendersMap[senderStr]
		if !exists {
			uniqueSendersMap[senderStr] = big.NewInt(0)
		}

		gasToUse := innerTx.GetGasLimit() - proc.economicsFee.ComputeGasLimit(innerTx)
		moveBalanceUserFee := proc.economicsFee.ComputeMoveBalanceFee(innerTx)
		processingUserFee := proc.economicsFee.ComputeFeeForProcessing(innerTx, gasToUse)
		innerTxFee := big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)

		uniqueSendersMap[senderStr].Add(uniqueSendersMap[senderStr], innerTxFee)
	}

	return uniqueSendersMap
}

func (proc *relayedTxV3Processor) computeRelayedTxMinGasLimit(tx *transaction.Transaction) uint64 {
	relayedTxGasLimit := proc.economicsFee.ComputeGasLimit(tx)
	uniqueSenders := proc.getUniqueSendersRequiredGasLimitsMap(tx.InnerTransactions)

	totalGasLimit := relayedTxGasLimit * uint64(len(uniqueSenders))
	for _, gasLimit := range uniqueSenders {
		totalGasLimit += gasLimit
	}

	return totalGasLimit
}

func (proc *relayedTxV3Processor) getUniqueSendersRequiredGasLimitsMap(innerTxs []*transaction.Transaction) map[string]uint64 {
	uniqueSendersMap := make(map[string]uint64)
	for _, innerTx := range innerTxs {
		senderStr := string(innerTx.SndAddr)
		_, exists := uniqueSendersMap[senderStr]
		if !exists {
			uniqueSendersMap[senderStr] = 0
		}

		uniqueSendersMap[senderStr] += innerTx.GasLimit
	}

	return uniqueSendersMap
}

// IsInterfaceNil returns true if there is no value under the interface
func (proc *relayedTxV3Processor) IsInterfaceNil() bool {
	return proc == nil
}
