package preprocess

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
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

// GetBalanceOfAddress method gets balance of an address
func (bc *balanceComputation) GetBalanceOfAddress(address []byte) (*big.Int, error) {
	bc.mutAddressBalance.RLock()
	defer bc.mutAddressBalance.RUnlock()

	currValue, ok := bc.mapAddressBalance[string(address)]
	if !ok {
		return nil, process.ErrAddressHasNoBalanceSet
	}

	return big.NewInt(0).Set(currValue), nil
}

// AddBalanceToAddress method adds balance to an address
func (bc *balanceComputation) AddBalanceToAddress(address []byte, value *big.Int) bool {
	bc.mutAddressBalance.Lock()
	defer bc.mutAddressBalance.Unlock()

	addedWithSuccess := false
	if currValue, ok := bc.mapAddressBalance[string(address)]; ok {
		currValue.Add(currValue, value)
		addedWithSuccess = true
	}

	return addedWithSuccess
}

// SubBalanceFromAddress method subtracts balance from an address
func (bc *balanceComputation) SubBalanceFromAddress(address []byte, value *big.Int) (bool, bool) {
	bc.mutAddressBalance.Lock()
	defer bc.mutAddressBalance.Unlock()

	subtractedWithSuccess := false
	addressExists := false
	if currValue, ok := bc.mapAddressBalance[string(address)]; ok {
		addressExists = true
		if currValue.Cmp(value) >= 0 {
			currValue.Sub(currValue, value)
			subtractedWithSuccess = true
		}
	}

	return addressExists, subtractedWithSuccess
}

// HasAddressBalanceSet method returns if address has balance set or not
func (bc *balanceComputation) HasAddressBalanceSet(address []byte) bool {
	bc.mutAddressBalance.RLock()
	defer bc.mutAddressBalance.RUnlock()

	_, ok := bc.mapAddressBalance[string(address)]
	return ok
}

// IsBalanceInAddress method returns if in given address is at least given balance
func (bc *balanceComputation) IsBalanceInAddress(address []byte, value *big.Int) bool {
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
