package postprocess

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TestIntermediateResProc extends intermediateResultsProcessor and is used in integration tests
// as it exposes some functions that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestIntermediateResProc struct {
	*intermediateResultsProcessor
}

// NewTestIntermediateResultsProcessor creates a new instance of TestIntermediateResProc
func NewTestIntermediateResultsProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	coordinator sharding.Coordinator,
	pubkeyConv core.PubkeyConverter,
	store dataRetriever.StorageService,
	blockType block.Type,
	currTxs dataRetriever.TransactionCacher,
	economicsFee process.FeeHandler,
) (*TestIntermediateResProc, error) {
	interimProc, err := NewIntermediateResultsProcessor(hasher, marshalizer, coordinator, pubkeyConv, store, blockType, currTxs, economicsFee)
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
