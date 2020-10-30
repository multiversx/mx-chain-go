package postprocess

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TestIntermediateResProc --
type TestIntermediateResProc struct {
	*intermediateResultsProcessor
}

// NewTestIntermediateResultsProcessor --
func NewTestIntermediateResultsProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	coordinator sharding.Coordinator,
	pubkeyConv core.PubkeyConverter,
	store dataRetriever.StorageService,
	blockType block.Type,
	currTxs dataRetriever.TransactionCacher,
) (*TestIntermediateResProc, error) {
	interimProc, err := NewIntermediateResultsProcessor(hasher, marshalizer, coordinator, pubkeyConv, store, blockType, currTxs)
	return &TestIntermediateResProc{interimProc}, err
}

// GetIntermediateTransactions --
func (tirp *TestIntermediateResProc) GetIntermediateTransactions() []data.TransactionHandler {
	tirp.mutInterResultsForBlock.Lock()
	defer tirp.mutInterResultsForBlock.Unlock()

	intermediateTxs := make([]data.TransactionHandler, 0)
	for _, val := range tirp.interResultsForBlock {
		intermediateTxs = append(intermediateTxs, val.tx)
	}

	return intermediateTxs
}

// CleanIntermediateTransactions --
func (tirp *TestIntermediateResProc) CleanIntermediateTransactions() {
	tirp.mutInterResultsForBlock.Lock()
	defer tirp.mutInterResultsForBlock.Unlock()

	tirp.interResultsForBlock = map[string]*txInfo{}
}
