package transactionAPI

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgAPITransactionProcessor is structure that store components that are needed to create an api transaction processor
type ArgAPITransactionProcessor struct {
	RoundDuration            uint64
	GenesisTime              time.Time
	Marshalizer              marshal.Marshalizer
	AddressPubKeyConverter   core.PubkeyConverter
	ShardCoordinator         sharding.Coordinator
	HistoryRepository        dblookupext.HistoryRepository
	StorageService           dataRetriever.StorageService
	DataPool                 dataRetriever.PoolsHolder
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	FeeComputer              feeComputer
	TxTypeHandler            process.TxTypeHandler
	LogsFacade               LogsFacade
	DataFieldParser          DataFieldParser
	ChainID                  string
}
