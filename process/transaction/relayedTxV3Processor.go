package transaction

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const minTransactionsAllowed = 1

// ArgRelayedTxV3Processor is the DTO used to create a new instance of relayedTxV3Processor
type ArgRelayedTxV3Processor struct {
	EconomicsFee           process.FeeHandler
	ShardCoordinator       sharding.Coordinator
	ArgsParser             process.ArgumentsParser
	MaxTransactionsAllowed int
}

type relayedTxV3Processor struct {
	economicsFee           process.FeeHandler
	shardCoordinator       sharding.Coordinator
	argsParser             process.ArgumentsParser
	maxTransactionsAllowed int
}

// NewRelayedTxV3Processor returns a new instance of relayedTxV3Processor
func NewRelayedTxV3Processor(args ArgRelayedTxV3Processor) (*relayedTxV3Processor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}
	return &relayedTxV3Processor{
		economicsFee:           args.EconomicsFee,
		shardCoordinator:       args.ShardCoordinator,
		maxTransactionsAllowed: args.MaxTransactionsAllowed,
		argsParser:             args.ArgsParser,
	}, nil
}

func checkArgs(args ArgRelayedTxV3Processor) error {
	if check.IfNil(args.EconomicsFee) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ArgsParser) {
		return process.ErrNilArgumentParser
	}
	if args.MaxTransactionsAllowed < minTransactionsAllowed {
		return fmt.Errorf("%w for MaxTransactionsAllowed, provided %d, min expected %d", process.ErrInvalidValue, args.MaxTransactionsAllowed, minTransactionsAllowed)
	}

	return nil
}

// CheckRelayedTx checks the relayed transaction and its inner transactions
func (proc *relayedTxV3Processor) CheckRelayedTx(tx *transaction.Transaction) error {
	if len(tx.InnerTransactions) > proc.maxTransactionsAllowed {
		return process.ErrRelayedTxV3TooManyInnerTransactions
	}
	if tx.GetValue().Cmp(big.NewInt(0)) != 0 {
		return process.ErrRelayedTxV3ZeroVal
	}
	if !bytes.Equal(tx.RcvAddr, tx.SndAddr) {
		return process.ErrRelayedTxV3SenderDoesNotMatchReceiver
	}
	if len(tx.Data) > 0 {
		funcName, _, err := proc.argsParser.ParseCallData(string(tx.Data))
		if err == nil && isRelayedTx(funcName) {
			return process.ErrMultipleRelayedTxTypesIsNotAllowed
		}
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
	feesForInnerTxs := proc.getTotalFeesRequiredForInnerTxs(tx.InnerTransactions)

	relayerMoveBalanceFee := proc.economicsFee.ComputeMoveBalanceFee(tx)
	relayerFee := big.NewInt(0).Mul(relayerMoveBalanceFee, big.NewInt(int64(len(tx.InnerTransactions))))

	totalFee := big.NewInt(0).Add(relayerFee, feesForInnerTxs)

	return relayerFee, totalFee
}

func (proc *relayedTxV3Processor) getTotalFeesRequiredForInnerTxs(innerTxs []*transaction.Transaction) *big.Int {
	totalFees := big.NewInt(0)
	for _, innerTx := range innerTxs {
		gasToUse := innerTx.GetGasLimit() - proc.economicsFee.ComputeGasLimit(innerTx)
		moveBalanceUserFee := proc.economicsFee.ComputeMoveBalanceFee(innerTx)
		processingUserFee := proc.economicsFee.ComputeFeeForProcessing(innerTx, gasToUse)
		innerTxFee := big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)

		totalFees.Add(totalFees, innerTxFee)
	}

	return totalFees
}

func (proc *relayedTxV3Processor) computeRelayedTxMinGasLimit(tx *transaction.Transaction) uint64 {
	relayedTxGasLimit := proc.economicsFee.ComputeGasLimit(tx)

	totalGasLimit := relayedTxGasLimit * uint64(len(tx.InnerTransactions))
	for _, innerTx := range tx.InnerTransactions {
		totalGasLimit += innerTx.GasLimit
	}

	return totalGasLimit
}

// IsInterfaceNil returns true if there is no value under the interface
func (proc *relayedTxV3Processor) IsInterfaceNil() bool {
	return proc == nil
}
