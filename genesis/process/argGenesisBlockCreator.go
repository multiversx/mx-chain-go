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
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
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

// ArgsGenesisBlockCreator holds the arguments which are needed to create a genesis block
type ArgsGenesisBlockCreator struct {
	GenesisTime          uint64
	StartEpochNum        uint32
	Data                 dataComponentsHandler
	Core                 coreComponentsHandler
	Accounts             state.AccountsAdapter
	ValidatorAccounts    state.AccountsAdapter
	InitialNodesSetup    genesis.InitialNodesHandler
	Economics            process.EconomicsDataHandler
	ShardCoordinator     sharding.Coordinator
	AccountsParser       genesis.AccountsParser
	SmartContractParser  genesis.InitialSmartContractParser
	GasSchedule          core.GasScheduleNotifier
	TxLogsProcessor      process.TransactionLogProcessor
	VirtualMachineConfig config.VirtualMachineConfig
	HardForkConfig       config.HardforkConfig
	TrieStorageManagers  map[string]common.StorageManager
	SystemSCConfig       config.SystemSmartContractsConfig
	RoundConfig          *config.RoundConfig
	EpochConfig          *config.EpochConfig
	WorkingDir           string
	BlockSignKeyGen      crypto.KeyGenerator
	ChainRunType         common.ChainRunType

	GenesisNodePrice *big.Int
	GenesisString    string
	// created components
	importHandler update.ImportHandler
}
