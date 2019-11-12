package external

import (
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SCQueryService defines how data should be get from a SC account
type SCQueryService interface {
	ExecuteQuery(query *smartContract.SCQuery) (*vmcommon.VMOutput, error)
	IsInterfaceNil() bool
}

// StatusMetricsHandler is the interface that defines what a node details handler/provider should do
type StatusMetricsHandler interface {
	StatusMetricsMap() (map[string]interface{}, error)
	IsInterfaceNil() bool
}
