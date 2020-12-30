package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ArgsGenesisBlockCreator holds the arguments which are needed to create a genesis block
type ArgsGenesisBlockCreator struct {
	GenesisTime              uint64
	StartEpochNum            uint32
	Accounts                 state.AccountsAdapter
	ValidatorAccounts        state.AccountsAdapter
	PubkeyConv               core.PubkeyConverter
	InitialNodesSetup        genesis.InitialNodesHandler
	Economics                process.EconomicsDataHandler
	ShardCoordinator         sharding.Coordinator
	Store                    dataRetriever.StorageService
	Blkc                     data.ChainHandler
	Marshalizer              marshal.Marshalizer
	SignMarshalizer          marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	DataPool                 dataRetriever.PoolsHolder
	AccountsParser           genesis.AccountsParser
	SmartContractParser      genesis.InitialSmartContractParser
	GasSchedule              core.GasScheduleNotifier
	TxLogsProcessor          process.TransactionLogProcessor
	VirtualMachineConfig     config.VirtualMachineConfig
	HardForkConfig           config.HardforkConfig
	TrieStorageManagers      map[string]data.StorageManager
	ChainID                  string
	SystemSCConfig           config.SystemSmartContractsConfig
	GeneralConfig            *config.GeneralSettingsConfig
	BlockSignKeyGen          crypto.KeyGenerator
	ImportStartHandler       update.ImportStartHandler
	WorkingDir               string
	GenesisNodePrice         *big.Int
	GenesisString            string
	// created components
	importHandler update.ImportHandler
}
