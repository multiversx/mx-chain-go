package postprocess

import (
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.TransactionFeeHandler = (*feeHandler)(nil)

type feeData struct {
	cost   *big.Int
	devFee *big.Int
}

var zero = big.NewInt(0)

type feeHandler struct {
	mut                sync.RWMutex
	mapDependentHashes map[string][]byte
	mapHashFee         map[string]*feeData
	accumulatedFees    *big.Int
	developerFees      *big.Int
}

// NewFeeAccumulator constructor for the fee accumulator
func NewFeeAccumulator() (*feeHandler, error) {
	f := &feeHandler{}
	f.accumulatedFees = big.NewInt(0)
	f.developerFees = big.NewInt(0)
	f.mapHashFee = make(map[string]*feeData)
	f.mapDependentHashes = make(map[string][]byte)
	return f, nil
}

// CreateBlockStarted does the cleanup before creating a new block
func (f *feeHandler) CreateBlockStarted(gasAndFees scheduled.GasAndFees) {
	f.mut.Lock()
	f.mapHashFee = make(map[string]*feeData)
	f.mapDependentHashes = make(map[string][]byte)
	f.accumulatedFees = big.NewInt(0)
	if gasAndFees.AccumulatedFees != nil {
		f.accumulatedFees = big.NewInt(0).Set(gasAndFees.AccumulatedFees)
	}

	f.developerFees = big.NewInt(0)
	if gasAndFees.DeveloperFees != nil {
		f.developerFees = big.NewInt(0).Set(gasAndFees.DeveloperFees)
	}
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
	developerFees := big.NewInt(0).Set(f.developerFees)
	f.mut.RUnlock()

	return developerFees
}

// ProcessTransactionFee adds the tx cost to the accumulated amount
func (f *feeHandler) ProcessTransactionFee(cost *big.Int, devFee *big.Int, txHash []byte) {
	f.mut.Lock()
	f.processTransactionFee(cost, devFee, txHash)
	f.mut.Unlock()
}

// ProcessTransactionFeeRelayedUserTx processes the transaction fee for the execution of a relayed user transaction
func (f *feeHandler) ProcessTransactionFeeRelayedUserTx(cost *big.Int, devFee *big.Int, userTxHash []byte, originalTxHash []byte) {
	f.mut.Lock()
	f.linkRelayedUserTxToOriginalTx(userTxHash, originalTxHash)
	f.processTransactionFee(cost, devFee, userTxHash)
	f.mut.Unlock()
}

// RevertFees reverts the accumulated fees for txHashes
func (f *feeHandler) RevertFees(txHashes [][]byte) {
	f.mut.Lock()
	defer f.mut.Unlock()

	for _, txHash := range txHashes {
		f.revertFeesForDependentTxHash(txHash)
		f.revertFee(txHash)
	}
}

func (f *feeHandler) linkRelayedUserTxToOriginalTx(userTxHash []byte, originalTxHash []byte) {
	f.mapDependentHashes[string(originalTxHash)] = userTxHash
}

func (f *feeHandler) revertFeesForDependentTxHash(txHash []byte) {
	linkedTxHash, ok := f.mapDependentHashes[string(txHash)]
	if !ok {
		return
	}
	f.revertFee(linkedTxHash)
	delete(f.mapDependentHashes, string(txHash))
}

func (f *feeHandler) revertFee(txHash []byte) {
	fee, ok := f.mapHashFee[string(txHash)]
	if !ok {
		return
	}
	f.developerFees.Sub(f.developerFees, fee.devFee)
	f.accumulatedFees.Sub(f.accumulatedFees, fee.cost)
	delete(f.mapHashFee, string(txHash))
}

func (f *feeHandler) processTransactionFee(cost *big.Int, devFee *big.Int, txHash []byte) {
	if cost == nil {
		log.Error("nil cost in ProcessTransactionFee", "error", process.ErrNilValue.Error())
		return
	}
	if cost.Cmp(zero) < 0 {
		log.Error("negative cost in ProcessTransactionFee", "error", process.ErrNegativeValue.Error())
		return
	}

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
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *feeHandler) IsInterfaceNil() bool {
	return f == nil
}
