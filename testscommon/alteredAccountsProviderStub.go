package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts/shared"
)

// AlteredAccountsProviderStub -
type AlteredAccountsProviderStub struct {
	ExtractAlteredAccountsFromPoolCalled func(txPool *outport.Pool, options shared.AlteredAccountsOptions) (map[string]*outport.AlteredAccount, error)
}

// ExtractAlteredAccountsFromPool -
func (a *AlteredAccountsProviderStub) ExtractAlteredAccountsFromPool(txPool *outport.Pool, options shared.AlteredAccountsOptions) (map[string]*outport.AlteredAccount, error) {
	if a.ExtractAlteredAccountsFromPoolCalled != nil {
		return a.ExtractAlteredAccountsFromPoolCalled(txPool, options)
	}

	return nil, nil
}

// IsInterfaceNil -
func (a *AlteredAccountsProviderStub) IsInterfaceNil() bool {
	return a == nil
}
