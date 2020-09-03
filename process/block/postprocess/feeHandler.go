package postprocess

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.TransactionFeeHandler = (*feeHandler)(nil)

type feeData struct {
	cost   *big.Int
	devFee *big.Int
}

type feeHandler struct {
	mut             sync.RWMutex
	mapHashFee      map[string]*feeData
	accumulatedFees *big.Int
	developerFees   *big.Int
}

// NewFeeAccumulator constructor for the fee accumulator
func NewFeeAccumulator() (*feeHandler, error) {
	f := &feeHandler{}
	f.accumulatedFees = big.NewInt(0)
	f.developerFees = big.NewInt(0)
	f.mapHashFee = make(map[string]*feeData)
	return f, nil
}

// CreateBlockStarted does the cleanup before creating a new block
func (f *feeHandler) CreateBlockStarted() {
	f.mut.Lock()
	f.mapHashFee = make(map[string]*feeData)
	f.accumulatedFees = big.NewInt(0)
	f.developerFees = big.NewInt(0)
	f.mut.Unlock()
}

// GetAccumulatedFees returns the total accumulated fees
func (f *feeHandler) GetAccumulatedFees() *big.Int {
	f.mut.RLock()
	accumulatedFees := big.NewInt(0).Set(f.accumulatedFees)
	f.mut.RUnlock()

	return accumulatedFees
}

// GetDeveloperFees returns the total accumulated developer fees
func (f *feeHandler) GetDeveloperFees() *big.Int {
	f.mut.RLock()
	developerFees := f.developerFees
	f.mut.RUnlock()

	return developerFees
}

// ProcessTransactionFee adds the tx cost to the accumulated amount
func (f *feeHandler) ProcessTransactionFee(cost *big.Int, devFee *big.Int, txHash []byte) {
	if cost == nil {
		log.Debug("nil cost in ProcessTransactionFee", "error", process.ErrNilValue.Error())
		return
	}

	f.mut.Lock()
	fee, ok := f.mapHashFee[string(txHash)]
	if !ok {
		fee = &feeData{
			cost:   big.NewInt(0),
			devFee: big.NewInt(0),
		}
	}

	fee.cost.Add(fee.cost, cost)
	fee.devFee.Add(fee.devFee, devFee)

	f.mapHashFee[string(txHash)] = fee
	f.accumulatedFees.Add(f.accumulatedFees, cost)
	f.developerFees.Add(f.developerFees, devFee)
	f.mut.Unlock()
}

// RevertFees reverts the accumulated fees for txHashes
func (f *feeHandler) RevertFees(txHashes [][]byte) {
	f.mut.Lock()
	defer f.mut.Unlock()

	for _, txHash := range txHashes {
		fee, ok := f.mapHashFee[string(txHash)]
		if !ok {
			continue
		}
		f.developerFees.Sub(f.developerFees, fee.devFee)
		f.accumulatedFees.Sub(f.accumulatedFees, fee.cost)
		delete(f.mapHashFee, string(txHash))
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *feeHandler) IsInterfaceNil() bool {
	return f == nil
}
