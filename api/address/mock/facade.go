package mock

import (
	"math/big"
)

type Facade struct {
	BalanceHandler             func(string) (*big.Int, error)
	GetCurrentPublicKeyHandler func() (string, error)
}

func (f *Facade) GetBalance(address string) (*big.Int, error) {
	return f.BalanceHandler(address)
}

func (f *Facade) GetCurrentPublicKey() (string, error) {
	return f.GetCurrentPublicKeyHandler()
}

type WrongFacade struct {
}
