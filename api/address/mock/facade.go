package mock

import (
	"math/big"
)

type Facade struct {
	BalanceHandler func(string) (*big.Int, error)
}

func (f *Facade) GetBalance(address string) (*big.Int, error) {
	return f.BalanceHandler(address)
}

type WrongFacade struct {
}
