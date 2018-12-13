package mock

import (
	"math/big"
)

type Facade struct {
	GetBalanceCalled func(string) (*big.Int, error)
}

func (f *Facade) GetBalance(address string) (*big.Int, error) {
	return f.GetBalanceCalled(address)
}

type WrongFacade struct {
}
