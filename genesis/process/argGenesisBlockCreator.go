package process

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/update"
)

type coreComponentsHandler interface {
	InternalMarshalizer() marshal.Marshalizer
	TxMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	AddressPubKeyConverter() core.PubkeyConverter
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	TxVersionChecker() process.TxVersionCheckerHandler
	ChainID() string
	EnableEpochsHandler() common.EnableEpochsHandler
	IsInterfaceNil() bool
}

type dataComponentsHandler interface {
	StorageService() dataRetriever.StorageService
	Blockchain() data.ChainHandler
	Datapool() dataRetriever.PoolsHolder
	SetBlockchain(chain data.ChainHandler) error
	Clone() interface{}
	IsInterfaceNil() bool
}

type runTypeComponentsHandler interface {
	BlockChainHookHandlerCreator() hooks.BlockChainHookHandlerCreator
	TransactionCoordinatorCreator() coordinator.TransactionCoordinatorCreator
	SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator
	SCProcessorCreator() scrCommon.SCProcessorCreator
	AccountsCreator() state.AccountFactory
	IsInterfaceNil() bool
}

// ArgsGenesisBlockCreator holds the arguments which are needed to create a genesis block
type ArgsGenesisBlockCreator struct {
	GenesisTime             uint64
	GenesisNonce            uint64
	GenesisRound            uint64
	StartEpochNum           uint32
	GenesisEpoch            uint32
	Data                    dataComponentsHandler
	Core                    coreComponentsHandler
	Accounts                state.AccountsAdapter
	ValidatorAccounts       state.AccountsAdapter
	InitialNodesSetup       genesis.InitialNodesHandler
	Economics               process.EconomicsDataHandler
	ShardCoordinator        sharding.Coordinator
	AccountsParser          genesis.AccountsParser
	SmartContractParser     genesis.InitialSmartContractParser
	GasSchedule             core.GasScheduleNotifier
	TxLogsProcessor         process.TransactionLogProcessor
	VirtualMachineConfig    config.VirtualMachineConfig
	HardForkConfig          config.HardforkConfig
	TrieStorageManagers     map[string]common.StorageManager
	SystemSCConfig          config.SystemSmartContractsConfig
	RoundConfig             config.RoundConfig
	EpochConfig             config.EpochConfig
	HeaderVersionConfigs    config.VersionsConfig
	WorkingDir              string
	BlockSignKeyGen         crypto.KeyGenerator
	HistoryRepository       dblookupext.HistoryRepository
	TxExecutionOrderHandler common.TxExecutionOrderHandler
	TxPreprocessorCreator   preprocess.TxPreProcessorCreator
	RunTypeComponents       runTypeComponentsHandler
	Config                  config.Config

	GenesisNodePrice *big.Int
	GenesisString    string

	// created components
	importHandler          update.ImportHandler
	versionedHeaderFactory genesis.VersionedHeaderFactory
	importHandler update.ImportHandler

	ShardCoordinatorFactory sharding.ShardCoordinatorFactory
	DNSV2Addresses          []string
}
