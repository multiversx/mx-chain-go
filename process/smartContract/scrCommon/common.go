package scrCommon

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// TestSmartContractProcessor is a SmartContractProcessor used in integration tests
type TestSmartContractProcessor interface {
	process.SmartContractProcessorFacade
	GetCompositeTestError() error
	GetGasRemaining() uint64
	GetAllSCRs() []data.TransactionHandler
	CleanGasRefunded()
}

// ArgsNewSmartContractProcessor defines the arguments needed for new smart contract processor
type ArgsNewSmartContractProcessor struct {
	VmContainer         process.VirtualMachinesContainer
	ArgsParser          process.ArgumentsParser
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	AccountsDB          state.AccountsAdapter
	BlockChainHook      process.BlockChainHookHandler
	BuiltInFunctions    vmcommon.BuiltInFunctionContainer
	PubkeyConv          core.PubkeyConverter
	ShardCoordinator    sharding.Coordinator
	ScrForwarder        process.IntermediateTransactionHandler
	TxFeeHandler        process.TransactionFeeHandler
	EconomicsFee        process.FeeHandler
	TxTypeHandler       process.TxTypeHandler
	GasHandler          process.GasHandler
	GasSchedule         core.GasScheduleNotifier
	TxLogsProcessor     process.TransactionLogProcessor
	BadTxForwarder      process.IntermediateTransactionHandler
	EnableRoundsHandler process.EnableRoundsHandler
	EnableEpochsHandler common.EnableEpochsHandler
	EnableEpochs        config.EnableEpochs
	VMOutputCacher      storage.Cacher
	WasmVMChangeLocker  common.Locker
	IsGenesisProcessing bool
}

// will uncomment when merging blhook_runsc

// FindVMByScAddress is exported for use in all version of scr processors
//lint:ignore U1000 Ignore unused function temporarily until merge
func FindVMByScAddress(container process.VirtualMachinesContainer, scAddress []byte) (vmcommon.VMExecutionHandler, []byte, error) {
	vmType, err := parseVMTypeFromContractAddress(scAddress)
	if err != nil {
		return nil, nil, err
	}

	vm, err := container.Get(vmType)
	if err != nil {
		return nil, nil, err
	}

	return vm, vmType, nil
}

func parseVMTypeFromContractAddress(contractAddress []byte) ([]byte, error) {
	if len(contractAddress) < core.NumInitCharactersForScAddress {
		return nil, process.ErrInvalidVMType
	}

	startIndex := core.NumInitCharactersForScAddress - core.VMTypeLen
	endIndex := core.NumInitCharactersForScAddress
	return contractAddress[startIndex:endIndex], nil
}
