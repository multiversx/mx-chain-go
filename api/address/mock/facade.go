package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// Facade is the mock implementation of a address router handler
type Facade struct {
	BalanceHandler    func(string) (*big.Int, error)
	GetAccountHandler func(address string) (*state.Account, error)
}

// GetBalance is the mock implementation of a handler's GetBalance method
func (f *Facade) GetBalance(address string) (*big.Int, error) {
	return f.BalanceHandler(address)
}

// GetAccount is the mock implementation of a handler's GetAccount method
func (f *Facade) GetAccount(address string) (*state.Account, error) {
	return f.GetAccountHandler(address)
}

// WrongFacade is a struct that can be used as a wrong implementation of the address router handler
type WrongFacade struct {
}
