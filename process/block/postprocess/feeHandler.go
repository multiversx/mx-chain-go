package postprocess

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

type feeHandler struct {
	mut             sync.RWMutex
	mapHashFee      map[string]*big.Int
	accumulatedFees *big.Int
}

// NewFeeAccumulator constructor for the fee accumulator
func NewFeeAccumulator() (*feeHandler, error) {
	f := &feeHandler{}
	f.accumulatedFees = big.NewInt(0)
	f.mapHashFee = make(map[string]*big.Int)
	return f, nil
}

// CreateBlockStarted does the cleanup before creating a new block
func (f *feeHandler) CreateBlockStarted() {
	f.mut.Lock()
	f.mapHashFee = make(map[string]*big.Int)
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
func (f *feeHandler) ProcessTransactionFee(cost *big.Int, txHash []byte) {
	if cost == nil {
		log.Debug("nil cost in ProcessTransactionFee", "error", process.ErrNilValue.Error())
		return
	}

	// TODO: Remove mutex, since all processing is performed sequentially?
	f.mut.Lock()
	f.mapHashFee[string(txHash)] = big.NewInt(0).Set(cost)
	f.accumulatedFees.Add(f.accumulatedFees, cost)
	f.mut.Unlock()
}

// RevertFees reverts the accumulated fees for txHashes
func (f *feeHandler) RevertFees(txHashes [][]byte) {
	f.mut.Lock()
	defer f.mut.Unlock()

	for _, txHash := range txHashes {
		cost, ok := f.mapHashFee[string(txHash)]
		if !ok {
			continue
		}

		f.accumulatedFees.Sub(f.accumulatedFees, cost)
		delete(f.mapHashFee, string(txHash))
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *feeHandler) IsInterfaceNil() bool {
	return f == nil
}
