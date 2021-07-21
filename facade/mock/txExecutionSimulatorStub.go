package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
)

// TxExecutionSimulatorStub -
type TxExecutionSimulatorStub struct {
	ProcessTxCalled func(tx *transaction.Transaction) (*data.SimulationResults, error)
}

// ProcessTx -
func (t *TxExecutionSimulatorStub) ProcessTx(tx *transaction.Transaction) (*data.SimulationResults, error) {
	if t.ProcessTxCalled != nil {
		return t.ProcessTxCalled(tx)
	}

	return &data.SimulationResults{}, nil
}

// IsInterfaceNil -
func (t *TxExecutionSimulatorStub) IsInterfaceNil() bool {
	return t == nil
}
