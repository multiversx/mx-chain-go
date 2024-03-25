package api

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/vm"
)

// SCQueryElementArgs -
type SCQueryElementArgs struct {
	GeneralConfig         *config.Config
	EpochConfig           *config.EpochConfig
	CoreComponents        factory.CoreComponentsHolder
	StateComponents       factory.StateComponentsHolder
	StatusCoreComponents  factory.StatusCoreComponentsHolder
	DataComponents        factory.DataComponentsHolder
	ProcessComponents     factory.ProcessComponentsHolder
	RunTypeComponents     factory.RunTypeComponentsHolder
	GasScheduleNotifier   core.GasScheduleNotifier
	MessageSigVerifier    vm.MessageSignVerifier
	SystemSCConfig        *config.SystemSmartContractsConfig
	Bootstrapper          process.Bootstrapper
	AllowVMQueriesChan    chan struct{}
	WorkingDir            string
	Index                 int
	GuardedAccountHandler process.GuardedAccountHandler
}

// CreateScQueryElement -
func CreateScQueryElement(args SCQueryElementArgs) (process.SCQueryService, common.StorageManager, error) {
	return createScQueryElement(scQueryElementArgs{
		generalConfig:         args.GeneralConfig,
		epochConfig:           args.EpochConfig,
		coreComponents:        args.CoreComponents,
		stateComponents:       args.StateComponents,
		dataComponents:        args.DataComponents,
		processComponents:     args.ProcessComponents,
		statusCoreComponents:  args.StatusCoreComponents,
		gasScheduleNotifier:   args.GasScheduleNotifier,
		messageSigVerifier:    args.MessageSigVerifier,
		systemSCConfig:        args.SystemSCConfig,
		bootstrapper:          args.Bootstrapper,
		allowVMQueriesChan:    args.AllowVMQueriesChan,
		workingDir:            args.WorkingDir,
		index:                 args.Index,
		guardedAccountHandler: args.GuardedAccountHandler,
		runTypeComponents:     args.RunTypeComponents,
	})
}

// CreateArgsSCQueryService -
func CreateArgsSCQueryService(args SCQueryElementArgs) (*smartContract.ArgsNewSCQueryService, error) {
	return createArgsSCQueryService(&scQueryElementArgs{
		generalConfig:         args.GeneralConfig,
		epochConfig:           args.EpochConfig,
		coreComponents:        args.CoreComponents,
		stateComponents:       args.StateComponents,
		dataComponents:        args.DataComponents,
		processComponents:     args.ProcessComponents,
		statusCoreComponents:  args.StatusCoreComponents,
		gasScheduleNotifier:   args.GasScheduleNotifier,
		messageSigVerifier:    args.MessageSigVerifier,
		systemSCConfig:        args.SystemSCConfig,
		bootstrapper:          args.Bootstrapper,
		allowVMQueriesChan:    args.AllowVMQueriesChan,
		workingDir:            args.WorkingDir,
		index:                 args.Index,
		guardedAccountHandler: args.GuardedAccountHandler,
		runTypeComponents:     args.RunTypeComponents,
	})
}

// CreateBlockchainForScQuery -
func CreateBlockchainForScQuery(selfShardID uint32) (data.ChainHandler, error) {
	return createBlockchainForScQuery(selfShardID)
}
