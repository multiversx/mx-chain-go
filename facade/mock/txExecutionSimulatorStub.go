package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
)

// TxExecutionSimulatorStub -
type TxExecutionSimulatorStub struct {
	ProcessTxCalled func(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
}

// ProcessTx -
func (t *TxExecutionSimulatorStub) ProcessTx(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
	if t.ProcessTxCalled != nil {
		return t.ProcessTxCalled(tx)
	}

	return &txSimData.SimulationResults{}, nil
}

// IsInterfaceNil -
func (t *TxExecutionSimulatorStub) IsInterfaceNil() bool {
	return t == nil
}
