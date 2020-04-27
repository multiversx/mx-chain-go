package process

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgsGenesisBlockCreator holds the arguments which are needed to create a genesis metablock
type ArgsGenesisBlockCreator struct {
	GenesisTime              uint64
	StartEpochNum            uint32
	Accounts                 state.AccountsAdapter
	PubkeyConv               state.PubkeyConverter
	InitialNodesSetup        genesis.InitialNodesHandler
	Economics                *economics.EconomicsData //TODO refactor and use an interface
	ShardCoordinator         sharding.Coordinator
	Store                    dataRetriever.StorageService
	Blkc                     data.ChainHandler
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	DataPool                 dataRetriever.PoolsHolder
	AccountsParser           genesis.AccountsParser
	ValidatorStatsRootHash   []byte
	GasMap                   map[string]map[string]uint64
	TxLogsProcessor          process.TransactionLogProcessor
	VirtualMachineConfig     config.VirtualMachineConfig
}
