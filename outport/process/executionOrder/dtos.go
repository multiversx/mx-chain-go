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
	transactionsToMe          []data.TransactionHandlerWithGasUsedAndFee
	scheduledTransactionsToMe []data.TransactionHandlerWithGasUsedAndFee
	scrsToMe                  map[string]data.TransactionHandlerWithGasUsedAndFee
}

type resultsTransactionsFromMe struct {
	transactionsFromMe                         []data.TransactionHandlerWithGasUsedAndFee
	scheduledTransactionsFromMe                []data.TransactionHandlerWithGasUsedAndFee
	scheduledExecutedInvalidTxsHashesPrevBlock []string
}
