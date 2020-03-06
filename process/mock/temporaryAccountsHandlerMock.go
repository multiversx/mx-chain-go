package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// TemporaryAccountsHandlerMock -
type TemporaryAccountsHandlerMock struct {
	AddTempAccountCalled    func(address []byte, balance *big.Int, nonce uint64)
	CleanTempAccountsCalled func()
	TempAccountCalled       func(address []byte) state.AccountHandler
}

// AddTempAccount -
func (tahm *TemporaryAccountsHandlerMock) AddTempAccount(address []byte, balance *big.Int, nonce uint64) {
	if tahm.AddTempAccountCalled == nil {
		return
	}

	tahm.AddTempAccountCalled(address, balance, nonce)
}

// CleanTempAccounts -
func (tahm *TemporaryAccountsHandlerMock) CleanTempAccounts() {
	if tahm.CleanTempAccountsCalled == nil {
		return
	}

	tahm.CleanTempAccountsCalled()
}

// TempAccount -
func (tahm *TemporaryAccountsHandlerMock) TempAccount(address []byte) state.AccountHandler {
	if tahm.TempAccountCalled == nil {
		return nil
	}

	return tahm.TempAccountCalled(address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tahm *TemporaryAccountsHandlerMock) IsInterfaceNil() bool {
	return tahm == nil
}
