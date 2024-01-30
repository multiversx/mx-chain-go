package vm

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type ArgsVmContainerFactory struct {
	Config              config.VirtualMachineConfig
	BlockGasLimit       uint64
	GasSchedule         core.GasScheduleNotifier
	EpochNotifier       process.EpochNotifier
	EnableEpochsHandler common.EnableEpochsHandler
	WasmVMChangeLocker  common.Locker
	ESDTTransferParser  vmcommon.ESDTTransferParser
	BuiltInFunctions    vmcommon.BuiltInFunctionContainer
	BlockChainHook      process.BlockChainHookWithAccountsAdapter
	Hasher              hashing.Hasher
	Economics           process.EconomicsDataHandler
	MessageSignVerifier vm.MessageSignVerifier
	NodesConfigProvider vm.NodesConfigProvider
	Marshalizer         marshal.Marshalizer
	SystemSCConfig      *config.SystemSmartContractsConfig
	ValidatorAccountsDB state.AccountsAdapter
	UserAccountsDB      state.AccountsAdapter
	ChanceComputer      nodesCoordinator.ChanceComputer
	ShardCoordinator    sharding.Coordinator
	PubkeyConv          core.PubkeyConverter
}
