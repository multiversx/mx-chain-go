package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxExecutionSimulatorStub -
type TxExecutionSimulatorStub struct {
	ProcessTxCalled func(tx *transaction.Transaction) (*transaction.SimulationResults, error)
}

// ProcessTx -
func (t *TxExecutionSimulatorStub) ProcessTx(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
	if t.ProcessTxCalled != nil {
		return t.ProcessTxCalled(tx)
	}

	return &transaction.SimulationResults{}, nil
}

// IsInterfaceNil -
func (t *TxExecutionSimulatorStub) IsInterfaceNil() bool {
	return t == nil
}
