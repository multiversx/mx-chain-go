package state

import (
	"math/big"

	"github.com/multiversx/mx-chain-vm-common-go"
)

// TODO: move all the mocks from the mock package to testscommon

// AccountsFactoryStub -
type AccountsFactoryStub struct {
	CreateAccountCalled func(address []byte) (vmcommon.AccountHandler, error)
}

// CreateAccount -
func (afs *AccountsFactoryStub) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	if afs.CreateAccountCalled != nil {
		return afs.CreateAccountCalled(address)
	}

	return &UserAccountStub{Address: address, Balance: big.NewInt(0)}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (afs *AccountsFactoryStub) IsInterfaceNil() bool {
	return afs == nil
}
