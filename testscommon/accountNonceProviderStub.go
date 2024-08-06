package testscommon

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/state"
)

// AccountNonceProviderStub -
type AccountNonceProviderStub struct {
	accountsAdapter state.AccountsAdapter

	GetAccountNonceCalled func(address []byte) (uint64, error)
}

func NewAccountNonceProviderStub() *AccountNonceProviderStub {
	return &AccountNonceProviderStub{}
}

// GetAccountNonce -
func (stub *AccountNonceProviderStub) GetAccountNonce(address []byte) (uint64, error) {
	if stub.GetAccountNonceCalled != nil {
		return stub.GetAccountNonceCalled(address)
	}

	if !check.IfNil(stub.accountsAdapter) {
		account, err := stub.accountsAdapter.GetExistingAccount(address)
		if err != nil {
			return 0, err
		}

		return account.GetNonce(), nil
	}

	return 0, errors.New("both AccountNonceProviderStub.GetAccountNonceCalled() and AccountNonceProviderStub.accountsAdapter are nil")
}

func (stub *AccountNonceProviderStub) SetAccountsAdapter(accountsAdapter state.AccountsAdapter) error {
	stub.accountsAdapter = accountsAdapter
	return nil
}

// IsInterfaceNil -
func (stub *AccountNonceProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
