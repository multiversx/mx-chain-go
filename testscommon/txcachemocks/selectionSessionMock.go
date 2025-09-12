package txcachemocks

import (
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// SelectionSessionMock -
type SelectionSessionMock struct {
	mutex            sync.Mutex
	accountByAddress map[string]*stateMock.UserAccountStub

	NumCallsGetAccountNonceAndBalance int
	GetAccountNonceAndBalanceCalled   func(address []byte) (uint64, *big.Int, bool, error)
	GetRootHashCalled                 func() ([]byte, error)
	IsIncorrectlyGuardedCalled        func(tx data.TransactionHandler) bool
}

// NewSelectionSessionMock -
func NewSelectionSessionMock() *SelectionSessionMock {
	return &SelectionSessionMock{
		accountByAddress: make(map[string]*stateMock.UserAccountStub),
	}
}

// NewSelectionSessionMockWithAccounts -
func NewSelectionSessionMockWithAccounts(accountsByAddress map[string]*stateMock.UserAccountStub) *SelectionSessionMock {
	return &SelectionSessionMock{
		accountByAddress: accountsByAddress,
	}
}

// SetNonce -
func (mock *SelectionSessionMock) SetNonce(address []byte, nonce uint64) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	key := string(address)

	if mock.accountByAddress[key] == nil {
		mock.accountByAddress[key] = newDefaultAccount()
	}

	mock.accountByAddress[key].Nonce = nonce
}

// SetBalance -
func (mock *SelectionSessionMock) SetBalance(address []byte, balance *big.Int) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	key := string(address)

	if mock.accountByAddress[key] == nil {
		mock.accountByAddress[key] = newDefaultAccount()
	}

	mock.accountByAddress[key].Balance = balance
}

// GetAccountNonceAndBalance -
func (mock *SelectionSessionMock) GetAccountNonceAndBalance(address []byte) (uint64, *big.Int, bool, error) {
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

// GetRootHash -
func (mock *SelectionSessionMock) GetRootHash() ([]byte, error) {
	if mock.GetRootHashCalled != nil {
		return mock.GetRootHashCalled()
	}
	return nil, nil
}

// IsIncorrectlyGuarded -
func (mock *SelectionSessionMock) IsIncorrectlyGuarded(tx data.TransactionHandler) bool {
	if mock.IsIncorrectlyGuardedCalled != nil {
		return mock.IsIncorrectlyGuardedCalled(tx)
	}

	return false
}

// IsInterfaceNil -
func (mock *SelectionSessionMock) IsInterfaceNil() bool {
	return mock == nil
}

func newDefaultAccount() *stateMock.UserAccountStub {
	return &stateMock.UserAccountStub{
		Nonce:   0,
		Balance: big.NewInt(1000000000000000000),
	}
}
