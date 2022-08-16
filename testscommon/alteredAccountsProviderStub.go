package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
)

// AlteredAccountsProviderStub -
type AlteredAccountsProviderStub struct {
	ExtractAlteredAccountsFromPoolCalled func(txPool *outport.Pool) (map[string]*outport.AlteredAccount, error)
}

// ExtractAlteredAccountsFromPool -
func (a *AlteredAccountsProviderStub) ExtractAlteredAccountsFromPool(txPool *outport.Pool) (map[string]*outport.AlteredAccount, error) {
	if a.ExtractAlteredAccountsFromPoolCalled != nil {
		return a.ExtractAlteredAccountsFromPoolCalled(txPool)
	}

	return nil, nil
}

// IsInterfaceNil -
func (a *AlteredAccountsProviderStub) IsInterfaceNil() bool {
	return a == nil
}
