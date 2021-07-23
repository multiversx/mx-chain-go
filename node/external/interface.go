package external

import (
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
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
	StatusMetricsMapWithoutP2P() map[string]interface{}
	StatusP2pMetricsMap() map[string]interface{}
	StatusMetricsWithoutP2PPrometheusString() string
	EconomicsMetrics() map[string]interface{}
	ConfigMetrics() map[string]interface{}
	EnableEpochsMetrics() map[string]interface{}
	NetworkMetrics() map[string]interface{}
	IsInterfaceNil() bool
}

// TransactionCostHandler defines the actions which should be handler by a transaction cost estimator
type TransactionCostHandler interface {
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	IsInterfaceNil() bool
}

// TotalStakedValueHandler defines the behavior of a component able to return total staked value
type TotalStakedValueHandler interface {
	GetTotalStakedValue() (*api.StakeValues, error)
	IsInterfaceNil() bool
}

// DirectStakedListHandler defines the behavior of a component able to return the direct stake list
type DirectStakedListHandler interface {
	GetDirectStakedList() ([]*api.DirectStakedValue, error)
	IsInterfaceNil() bool
}

// DelegatedListHandler defines the behavior of a component able to return the complete delegated list
type DelegatedListHandler interface {
	GetDelegatorsList() ([]*api.Delegator, error)
	IsInterfaceNil() bool
}
