package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type accountStateProvider struct {
	accountsAdapter state.AccountsAdapter
	guardianChecker process.GuardianChecker
}

func newAccountStateProvider(accountsAdapter state.AccountsAdapter, guardianChecker process.GuardianChecker) (*accountStateProvider, error) {
	if check.IfNil(accountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(guardianChecker) {
		return nil, process.ErrNilGuardianChecker
	}

	return &accountStateProvider{
		accountsAdapter: accountsAdapter,
		guardianChecker: guardianChecker,
	}, nil
}

// GetAccountState returns the state of an account.
// Will be called by mempool during transaction selection.
func (provider *accountStateProvider) GetAccountState(address []byte) (*txcache.AccountState, error) {
	account, err := provider.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, errors.ErrWrongTypeAssertion
	}

	guardian, err := provider.getGuardian(userAccount)
	if err != nil {
		return nil, err
	}

	return &txcache.AccountState{
		Nonce:    userAccount.GetNonce(),
		Balance:  userAccount.GetBalance(),
		Guardian: guardian,
	}, nil
}

func (provider *accountStateProvider) getGuardian(userAccount state.UserAccountHandler) ([]byte, error) {
	if !userAccount.IsGuarded() {
		return nil, nil
	}

	vmUserAccount, ok := userAccount.(vmcommon.UserAccountHandler)
	if !ok {
		return nil, errors.ErrWrongTypeAssertion
	}

	return provider.guardianChecker.GetActiveGuardian(vmUserAccount)
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *accountStateProvider) IsInterfaceNil() bool {
	return provider == nil
}
