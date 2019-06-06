package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type FakeAccountsHandlerMock struct {
	CreateFakeAccountsCalled func(address []byte, balance *big.Int, nonce uint64)
	CleanFakeAccountsCalled  func()
	GetFakeAccountCalled     func(address []byte) state.AccountHandler
}

func (fake *FakeAccountsHandlerMock) CreateFakeAccounts(address []byte, balance *big.Int, nonce uint64) {
	if fake.CreateFakeAccountsCalled == nil {
		return
	}

	fake.CreateFakeAccountsCalled(address, balance, nonce)
}

func (fake *FakeAccountsHandlerMock) CleanFakeAccounts() {
	if fake.CleanFakeAccountsCalled == nil {
		return
	}

	fake.CleanFakeAccountsCalled()
}

func (fake *FakeAccountsHandlerMock) GetFakeAccount(address []byte) state.AccountHandler {
	if fake.GetFakeAccountCalled == nil {
		return nil
	}

	return fake.GetFakeAccountCalled(address)
}
