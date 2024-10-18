package transactionsfee

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

// FeesProcessorHandler defines the interface for the transaction fees processor
type FeesProcessorHandler interface {
	ComputeRelayedTxV3GasUnits(tx data.TransactionWithFeeHandler, epoch uint32) uint64
	ComputeRelayedTxFees(tx data.TransactionWithFeeHandler) (*big.Int, *big.Int, error)
	ComputeGasUnitsFromRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) uint64
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
	IsInterfaceNil() bool
}

type transactionGetter interface {
	GetTxByHash(txHash []byte) (*transaction.Transaction, error)
}

type dataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}
