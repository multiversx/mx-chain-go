package postprocess

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// TestIntermediateResProc extends intermediateResultsProcessor and is used in integration tests
// as it exposes some functions that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestIntermediateResProc struct {
	*intermediateResultsProcessor
}

// NewTestIntermediateResultsProcessor creates a new instance of TestIntermediateResProc
func NewTestIntermediateResultsProcessor(
	args ArgsNewIntermediateResultsProcessor,
) (*TestIntermediateResProc, error) {
	interimProc, err := NewIntermediateResultsProcessor(args)
	return &TestIntermediateResProc{interimProc}, err
}

// GetIntermediateTransactions returns all the intermediate transactions from the underlying map
func (tirp *TestIntermediateResProc) GetIntermediateTransactions() []data.TransactionHandler {
	tirp.mutInterResultsForBlock.Lock()
	defer tirp.mutInterResultsForBlock.Unlock()

	intermediateTxs := make([]data.TransactionHandler, 0)
	for _, val := range tirp.interResultsForBlock {
		intermediateTxs = append(intermediateTxs, val.tx)
	}

	return intermediateTxs
}

// CleanIntermediateTransactions removes the intermediate transactions from the underlying map
func (tirp *TestIntermediateResProc) CleanIntermediateTransactions() {
	tirp.mutInterResultsForBlock.Lock()
	defer tirp.mutInterResultsForBlock.Unlock()

	tirp.interResultsForBlock = map[string]*txInfo{}
}
