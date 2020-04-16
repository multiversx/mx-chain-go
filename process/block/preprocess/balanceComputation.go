package preprocess

import (
	"math/big"
	"sync"
)

type balanceComputation struct {
	mapAddressBalance map[string]*big.Int
	mutAddressBalance sync.RWMutex
}

// NewBalanceComputation creates a new object which computes the addresses balances
func NewBalanceComputation() (*balanceComputation, error) {
	return &balanceComputation{
		mapAddressBalance: make(map[string]*big.Int),
	}, nil
}

// Init method resets all the addresses balances
func (bc *balanceComputation) Init() {
	bc.mutAddressBalance.Lock()
	bc.mapAddressBalance = make(map[string]*big.Int)
	bc.mutAddressBalance.Unlock()
}

// SetBalanceToAddress method sets balance to an address
func (bc *balanceComputation) SetBalanceToAddress(address []byte, value *big.Int) {
	bc.mutAddressBalance.Lock()
	bc.mapAddressBalance[string(address)] = big.NewInt(0).Set(value)
	bc.mutAddressBalance.Unlock()
}

// AddBalanceToAddress method adds balance to an address
func (bc *balanceComputation) AddBalanceToAddress(address []byte, value *big.Int) bool {
	bc.mutAddressBalance.Lock()
	defer bc.mutAddressBalance.Unlock()

	if currValue, ok := bc.mapAddressBalance[string(address)]; ok {
		currValue.Add(currValue, value)
		return true
	}

	return false
}

// SubBalanceFromAddress method subtracts balance from an address
func (bc *balanceComputation) SubBalanceFromAddress(address []byte, value *big.Int) bool {
	bc.mutAddressBalance.Lock()
	defer bc.mutAddressBalance.Unlock()

	if currValue, ok := bc.mapAddressBalance[string(address)]; ok {
		if currValue.Cmp(value) >= 0 {
			currValue.Sub(currValue, value)
			return true
		}
	}

	return false
}

// IsAddressSet method returns if address is set
func (bc *balanceComputation) IsAddressSet(address []byte) bool {
	bc.mutAddressBalance.RLock()
	defer bc.mutAddressBalance.RUnlock()

	_, ok := bc.mapAddressBalance[string(address)]
	return ok
}

// AddressHasEnoughBalance method returns if given address has enough balance in
func (bc *balanceComputation) AddressHasEnoughBalance(address []byte, value *big.Int) bool {
	bc.mutAddressBalance.RLock()
	defer bc.mutAddressBalance.RUnlock()

	balance, ok := bc.mapAddressBalance[string(address)]
	if !ok {
		return false
	}

	return balance.Cmp(value) >= 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (bc *balanceComputation) IsInterfaceNil() bool {
	return bc == nil
}
