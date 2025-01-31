package blockAPI

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction/status"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	outportProcess "github.com/multiversx/mx-chain-go/outport/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

// ArgAPIBlockProcessor is structure that store components that are needed to create an api block processor
type ArgAPIBlockProcessor struct {
	SelfShardID                  uint32
	Store                        dataRetriever.StorageService
	Marshalizer                  marshal.Marshalizer
	Uint64ByteSliceConverter     typeConverters.Uint64ByteSliceConverter
	HistoryRepo                  dblookupext.HistoryRepository
	APITransactionHandler        APITransactionHandler
	StatusComputer               status.StatusComputerHandler
	Hasher                       hashing.Hasher
	AddressPubkeyConverter       core.PubkeyConverter
	LogsFacade                   logsFacade
	ReceiptsRepository           receiptsRepository
	AlteredAccountsProvider      outportProcess.AlteredAccountsProviderHandler
	AccountsRepository           state.AccountsRepository
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	EnableEpochsHandler          common.EnableEpochsHandler
}
