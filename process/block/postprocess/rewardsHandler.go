package postprocess

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

type feeHandler struct {
	mut             sync.Mutex
	accumulatedFees *big.Int
}

// NewFeeAccumulator constructor for the fee accumulator
func NewFeeAccumulator() (*feeHandler, error) {
	rtxh := &feeHandler{}
	rtxh.accumulatedFees = big.NewInt(0)
	return rtxh, nil
}

// CreateBlockStarted does the cleanup before creating a new block
func (rtxh *feeHandler) CreateBlockStarted() {
	rtxh.mut.Lock()
	rtxh.accumulatedFees = big.NewInt(0)
	rtxh.mut.Unlock()
}

// GetAccumulatedFees returns the total accumulated fees
func (rtxh *feeHandler) GetAccumulatedFees() *big.Int {
	rtxh.mut.Lock()
	accumulatedFees := big.NewInt(0).Set(rtxh.accumulatedFees)
	rtxh.mut.Unlock()

	return accumulatedFees
}

// ProcessTransactionFee adds the tx cost to the accumulated amount
func (rtxh *feeHandler) ProcessTransactionFee(cost *big.Int) {
	if cost == nil {
		log.Debug("nil cost in ProcessTransactionFee", "error", process.ErrNilValue.Error())
		return
	}

	rtxh.mut.Lock()
	_ = rtxh.accumulatedFees.Add(rtxh.accumulatedFees, cost)
	rtxh.mut.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtxh *feeHandler) IsInterfaceNil() bool {
	return rtxh == nil
}
