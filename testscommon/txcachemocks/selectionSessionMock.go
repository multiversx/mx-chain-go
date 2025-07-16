package txcachemocks

import (
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// SelectionSessionMock -
type SelectionSessionMock struct {
	mutex sync.Mutex

	NumCallsGetAccountState int

	AccountStateByAddress map[string]*stateMock.UserAccountStub
	GetAccountStateCalled func(address []byte) (state.UserAccountHandler, error)
	GetRootHashCalled     func() ([]byte, error)

	IsIncorrectlyGuardedCalled func(tx data.TransactionHandler) bool
}

// NewSelectionSessionMock -
func NewSelectionSessionMock() *SelectionSessionMock {
	return &SelectionSessionMock{
		AccountStateByAddress: make(map[string]*stateMock.UserAccountStub),
	}
}

// SetNonce -
func (mock *SelectionSessionMock) SetNonce(address []byte, nonce uint64) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	key := string(address)

	if mock.AccountStateByAddress[key] == nil {
		mock.AccountStateByAddress[key] = newDefaultAccountState()
	}

	mock.AccountStateByAddress[key].Nonce = nonce
}

// SetBalance -
func (mock *SelectionSessionMock) SetBalance(address []byte, balance *big.Int) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	key := string(address)

	if mock.AccountStateByAddress[key] == nil {
		mock.AccountStateByAddress[key] = newDefaultAccountState()
	}

	mock.AccountStateByAddress[key].Balance = balance
}

// GetAccountState -
func (mock *SelectionSessionMock) GetAccountState(address []byte) (state.UserAccountHandler, error) {
	mock.mutex.Lock()
	defer mock.mutex.Unlock()

	mock.NumCallsGetAccountState++

	if mock.GetAccountStateCalled != nil {
		return mock.GetAccountStateCalled(address)
	}

	state, ok := mock.AccountStateByAddress[string(address)]
	if ok {
		return state, nil
	}

	return newDefaultAccountState(), nil
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

func newDefaultAccountState() *stateMock.UserAccountStub {
	return &stateMock.UserAccountStub{
		Nonce:   0,
		Balance: big.NewInt(1000000000000000000),
	}
}
