package executionOrder

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
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
