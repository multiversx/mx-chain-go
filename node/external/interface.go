package external

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SCQueryService defines how data should be get from a SC account
type SCQueryService interface {
	ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error)
	Close() error
	IsInterfaceNil() bool
}

// StatusMetricsHandler is the interface that defines what a node details handler/provider should do
type StatusMetricsHandler interface {
	StatusMetricsMapWithoutP2P() (map[string]interface{}, error)
	StatusP2pMetricsMap() (map[string]interface{}, error)
	StatusMetricsWithoutP2PPrometheusString() (string, error)
	EconomicsMetrics() (map[string]interface{}, error)
	ConfigMetrics() (map[string]interface{}, error)
	EnableEpochsMetrics() (map[string]interface{}, error)
	NetworkMetrics() (map[string]interface{}, error)
	RatingsMetrics() (map[string]interface{}, error)
	IsInterfaceNil() bool
}

// TransactionCostHandler defines the actions which should be handler by a transaction cost estimator
type TransactionCostHandler interface {
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	IsInterfaceNil() bool
}

// TotalStakedValueHandler defines the behavior of a component able to return total staked value
type TotalStakedValueHandler interface {
	GetTotalStakedValue(ctx context.Context) (*api.StakeValues, error)
	IsInterfaceNil() bool
}

// DirectStakedListHandler defines the behavior of a component able to return the direct stake list
type DirectStakedListHandler interface {
	GetDirectStakedList(ctx context.Context) ([]*api.DirectStakedValue, error)
	IsInterfaceNil() bool
}

// DelegatedListHandler defines the behavior of a component able to return the complete delegated list
type DelegatedListHandler interface {
	GetDelegatorsList(ctx context.Context) ([]*api.Delegator, error)
	IsInterfaceNil() bool
}

// APITransactionHandler defines what an API transaction handler should be able to do
type APITransactionHandler interface {
	GetTransaction(txHash string, withResults bool) (*transaction.ApiTransactionResult, error)
	GetTransactionsPool() (*common.TransactionsPoolAPIResponse, error)
	GetTransactionsForSender(sender string) (*common.TransactionsForSenderApiResponse, error)
	UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	UnmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error)
	IsInterfaceNil() bool
}
