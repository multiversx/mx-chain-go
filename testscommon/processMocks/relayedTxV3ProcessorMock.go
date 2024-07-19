package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// RelayedTxV3ProcessorMock -
type RelayedTxV3ProcessorMock struct {
	CheckRelayedTxCalled func(tx *transaction.Transaction) error
}

// CheckRelayedTx -
func (mock *RelayedTxV3ProcessorMock) CheckRelayedTx(tx *transaction.Transaction) error {
	if mock.CheckRelayedTxCalled != nil {
		return mock.CheckRelayedTxCalled(tx)
	}
	return nil
}

// IsInterfaceNil -
func (mock *RelayedTxV3ProcessorMock) IsInterfaceNil() bool {
	return mock == nil
}
