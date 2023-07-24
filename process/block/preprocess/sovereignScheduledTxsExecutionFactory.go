package preprocess

import (
	processDisabled "github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignScheduledTxsExecutionFactory struct {
}

// NewSovereignScheduledTxsExecutionFactory creates a new sovereign scheduled txs execution factory
func NewSovereignScheduledTxsExecutionFactory() (*sovereignScheduledTxsExecutionFactory, error) {
	return &sovereignScheduledTxsExecutionFactory{}, nil
}

// CreateScheduledTxsExecutionHandler creates a new scheduled txs execution handler for sovereign chain
func (stxef *sovereignScheduledTxsExecutionFactory) CreateScheduledTxsExecutionHandler(_ ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error) {
	return &processDisabled.ScheduledTxsExecutionHandler{}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stxef *sovereignScheduledTxsExecutionFactory) IsInterfaceNil() bool {
	return stxef == nil
}
