package txcachemocks

import (
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// AccountNonceAndBalanceProviderMock -
type AccountNonceAndBalanceProviderMock struct {
	mutex            sync.Mutex
	accountByAddress map[string]*stateMock.UserAccountStub

	NumCallsGetAccountNonceAndBalance int
	GetAccountNonceAndBalanceCalled   func(address []byte) (uint64, *big.Int, bool, error)
}

// NewAccountNonceAndBalanceProviderMock -
func NewAccountNonceAndBalanceProviderMock() *AccountNonceAndBalanceProviderMock {
	return &AccountNonceAndBalanceProviderMock{
		accountByAddress: make(map[string]*stateMock.UserAccountStub),
	}
}

// NewAccountNonceAndBalanceProviderMockWithAccounts -
func NewAccountNonceAndBalanceProviderMockWithAccounts(accountsByAddress map[string]*stateMock.UserAccountStub) *AccountNonceAndBalanceProviderMock {
	return &AccountNonceAndBalanceProviderMock{
		accountByAddress: accountsByAddress,
	}
}

// SetNonce -
func (mock *AccountNonceAndBalanceProviderMock) SetNonce(address []byte, nonce uint64) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	key := string(address)

	if mock.accountByAddress[key] == nil {
		mock.accountByAddress[key] = newDefaultAccount()
	}

	mock.accountByAddress[key].Nonce = nonce
}

// SetBalance -
func (mock *AccountNonceAndBalanceProviderMock) SetBalance(address []byte, balance *big.Int) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	key := string(address)

	if mock.accountByAddress[key] == nil {
		mock.accountByAddress[key] = newDefaultAccount()
	}

	mock.accountByAddress[key].Balance = balance
}

// GetAccountNonceAndBalance -
func (mock *AccountNonceAndBalanceProviderMock) GetAccountNonceAndBalance(address []byte) (uint64, *big.Int, bool, error) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	mock.NumCallsGetAccountNonceAndBalance++

	if mock.GetAccountNonceAndBalanceCalled != nil {
		return mock.GetAccountNonceAndBalanceCalled(address)
	}

	account, ok := mock.accountByAddress[string(address)]
	if ok {
		if check.IfNil(account) {
			// This mock allows one to add "nil" (unknown) accounts in "AccountByAddress".
			return 0, big.NewInt(0), false, nil
		}

		return account.Nonce, account.Balance, true, nil
	}

	account = newDefaultAccount()
	return account.Nonce, account.Balance, true, nil
}

// IsInterfaceNil -
func (mock *AccountNonceAndBalanceProviderMock) IsInterfaceNil() bool {
	return mock == nil
}
