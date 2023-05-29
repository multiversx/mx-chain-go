package executionOrder

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

// ArgSorter holds the arguments needed for creating a new instance of sorter
type ArgSorter struct {
	Hasher              hashing.Hasher
	Marshaller          marshal.Marshalizer
	MbsStorer           storage.Storer
	EnableEpochsHandler common.EnableEpochsHandler
}

type resultsTransactionsToMe struct {
	transactionsToMe          []data.TxWithExecutionOrderHandler
	scheduledTransactionsToMe []data.TxWithExecutionOrderHandler
	scrsToMe                  map[string]data.TxWithExecutionOrderHandler
}

type resultsTransactionsFromMe struct {
	transactionsFromMe                         []data.TxWithExecutionOrderHandler
	scheduledTransactionsFromMe                []data.TxWithExecutionOrderHandler
	scheduledExecutedInvalidTxsHashesPrevBlock []string
}
