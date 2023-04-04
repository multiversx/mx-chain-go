package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts/shared"
)

// AlteredAccountsProviderStub -
type AlteredAccountsProviderStub struct {
	ExtractAlteredAccountsFromPoolCalled func(txPool *outport.TransactionPool, options shared.AlteredAccountsOptions) (map[string]*alteredAccount.AlteredAccount, error)
}

// ExtractAlteredAccountsFromPool -
func (a *AlteredAccountsProviderStub) ExtractAlteredAccountsFromPool(txPool *outport.TransactionPool, options shared.AlteredAccountsOptions) (map[string]*alteredAccount.AlteredAccount, error) {
	if a.ExtractAlteredAccountsFromPoolCalled != nil {
		return a.ExtractAlteredAccountsFromPoolCalled(txPool, options)
	}

	return nil, nil
}

// IsInterfaceNil -
func (a *AlteredAccountsProviderStub) IsInterfaceNil() bool {
	return a == nil
}
