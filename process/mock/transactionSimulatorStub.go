package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TransactionSimulatorStub -
type TransactionSimulatorStub struct {
	ProcessTxCalled func(tx *transaction.Transaction) (*transaction.SimulationResults, error)
}

// ProcessTx -
func (tss *TransactionSimulatorStub) ProcessTx(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
	if tss.ProcessTxCalled != nil {
		return tss.ProcessTxCalled(tx)
	}

	return nil, nil
}

// IsInterfaceNil -
func (tss *TransactionSimulatorStub) IsInterfaceNil() bool {
	return tss == nil
}
