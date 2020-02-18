package postprocess

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

type feeHandler struct {
	mut             sync.RWMutex
	accumulatedFees *big.Int
}

// NewFeeAccumulator constructor for the fee accumulator
func NewFeeAccumulator() (*feeHandler, error) {
	f := &feeHandler{}
	f.accumulatedFees = big.NewInt(0)
	return f, nil
}

// CreateBlockStarted does the cleanup before creating a new block
func (f *feeHandler) CreateBlockStarted() {
	f.mut.Lock()
	f.accumulatedFees = big.NewInt(0)
	f.mut.Unlock()
}

// GetAccumulatedFees returns the total accumulated fees
func (f *feeHandler) GetAccumulatedFees() *big.Int {
	f.mut.RLock()
	accumulatedFees := big.NewInt(0).Set(f.accumulatedFees)
	f.mut.RUnlock()

	return accumulatedFees
}

// ProcessTransactionFee adds the tx cost to the accumulated amount
func (f *feeHandler) ProcessTransactionFee(cost *big.Int) {
	if cost == nil {
		log.Debug("nil cost in ProcessTransactionFee", "error", process.ErrNilValue.Error())
		return
	}

	f.mut.Lock()
	f.accumulatedFees.Add(f.accumulatedFees, cost)
	f.mut.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *feeHandler) IsInterfaceNil() bool {
	return f == nil
}
