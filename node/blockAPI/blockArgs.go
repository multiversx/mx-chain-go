package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// APIBlockProcessorArg is structure that store components that are needed to create an api block procesosr
type APIBlockProcessorArg struct {
	SelfShardID              uint32
	Store                    dataRetriever.StorageService
	Marshalizer              marshal.Marshalizer
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	HistoryRepo              dblookupext.HistoryRepository
	UnmarshalTx              func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	StatusComputer           transaction.StatusComputerHandler
}
