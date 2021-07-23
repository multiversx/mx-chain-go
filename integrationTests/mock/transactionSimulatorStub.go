package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
)

// TransactionSimulatorStub -
type TransactionSimulatorStub struct {
	ProcessTxCalled func(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
}

// ProcessTx -
func (tss *TransactionSimulatorStub) ProcessTx(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
	if tss.ProcessTxCalled != nil {
		return tss.ProcessTxCalled(tx)
	}

	return nil, nil
}

// IsInterfaceNil -
func (tss *TransactionSimulatorStub) IsInterfaceNil() bool {
	return tss == nil
}
