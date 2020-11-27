package external

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

// SCQueryService defines how data should be get from a SC account
type SCQueryService interface {
	ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error)
	IsInterfaceNil() bool
}

// StatusMetricsHandler is the interface that defines what a node details handler/provider should do
type StatusMetricsHandler interface {
	StatusMetricsMapWithoutP2P() map[string]interface{}
	StatusP2pMetricsMap() map[string]interface{}
	StatusMetricsWithoutP2PPrometheusString() string
	EconomicsMetrics() map[string]interface{}
	ConfigMetrics() map[string]interface{}
	NetworkMetrics() map[string]interface{}
	IsInterfaceNil() bool
}

// TransactionCostHandler defines the actions which should be handler by a transaction cost estimator
type TransactionCostHandler interface {
	ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error)
	IsInterfaceNil() bool
}
