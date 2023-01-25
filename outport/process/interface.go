package process

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts/shared"
)

// AlteredAccountsProviderHandler defines the functionality needed for provisioning of altered accounts when indexing data
type AlteredAccountsProviderHandler interface {
	ExtractAlteredAccountsFromPool(txPool *outport.Pool, options shared.AlteredAccountsOptions) (map[string]*outport.AlteredAccount, error)
	IsInterfaceNil() bool
}

// TransactionsFeeHandler defines the functionality needed for computation of the transaction fee and gas used
type TransactionsFeeHandler interface {
	PutFeeAndGasUsed(pool *outport.Pool) error
	IsInterfaceNil() bool
}

// GasConsumedProvider defines the functionality needed for providing gas consumed information
type GasConsumedProvider interface {
	TotalGasProvided() uint64
	TotalGasProvidedWithScheduled() uint64
	TotalGasRefunded() uint64
	TotalGasPenalized() uint64
	IsInterfaceNil() bool
}

// EconomicsDataHandler defines the functionality needed for economics data
type EconomicsDataHandler interface {
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
	IsInterfaceNil() bool
	MaxGasLimitPerBlock(shardID uint32) uint64
}

// ExecutionOrderHandler defines the interface for the execution order handler
type ExecutionOrderHandler interface {
	PutExecutionOrderInTransactionPool(
		pool *outport.Pool,
		header data.HeaderHandler,
		body data.BodyHandler,
		prevHeader data.HeaderHandler,
	) ([]string, []string, error)
}
