package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common/dblookupext"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
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
