package testscommon

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/state"
)

type accountNonceProviderMock struct {
	accountsAdapter state.AccountsAdapter

	GetAccountNonceCalled func(address []byte) (uint64, error)
}

// NewAccountNonceProviderMock -
func NewAccountNonceProviderMock() *accountNonceProviderMock {
	return &accountNonceProviderMock{}
}

// GetAccountNonce -
func (stub *accountNonceProviderMock) GetAccountNonce(address []byte) (uint64, error) {
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

	return 0, errors.New("both accountNonceProviderStub.GetAccountNonceCalled() and accountNonceProviderStub.accountsAdapter are nil")
}

// SetAccountsAdapter -
func (stub *accountNonceProviderMock) SetAccountsAdapter(accountsAdapter state.AccountsAdapter) error {
	stub.accountsAdapter = accountsAdapter
	return nil
}

// IsInterfaceNil -
func (stub *accountNonceProviderMock) IsInterfaceNil() bool {
	return stub == nil
}
