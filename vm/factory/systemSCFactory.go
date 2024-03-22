package factory

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("vm/factory")

type systemSCFactory struct {
	systemEI               vm.ContextHandler
	economics              vm.EconomicsHandler
	nodesConfigProvider    vm.NodesConfigProvider
	sigVerifier            vm.MessageSignVerifier
	gasCost                vm.GasCost
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	systemSCConfig         *config.SystemSmartContractsConfig
	systemSCsContainer     vm.SystemSCContainer
	addressPubKeyConverter core.PubkeyConverter
	shardCoordinator       sharding.Coordinator
	enableEpochsHandler    common.EnableEpochsHandler
	nodesCoordinator       vm.NodesCoordinator
}

// ArgsNewSystemSCFactory defines the arguments struct needed to create the system SCs
type ArgsNewSystemSCFactory struct {
	SystemEI               vm.ContextHandler
	Economics              vm.EconomicsHandler
	NodesConfigProvider    vm.NodesConfigProvider
	SigVerifier            vm.MessageSignVerifier
	GasSchedule            core.GasScheduleNotifier
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	SystemSCConfig         *config.SystemSmartContractsConfig
	AddressPubKeyConverter core.PubkeyConverter
	ShardCoordinator       sharding.Coordinator
	EnableEpochsHandler    common.EnableEpochsHandler
	NodesCoordinator       vm.NodesCoordinator
}

// NewSystemSCFactory creates a factory which will instantiate the system smart contracts
func NewSystemSCFactory(args ArgsNewSystemSCFactory) (*systemSCFactory, error) {
	if check.IfNil(args.SystemEI) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilSystemEnvironmentInterface)
	}
	if check.IfNil(args.SigVerifier) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilMessageSignVerifier)
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilNodesConfigProvider)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilMarshalizer)
	}
	if check.IfNil(args.Hasher) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilHasher)
	}
	if check.IfNil(args.Economics) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilEconomicsData)
	}
	if args.SystemSCConfig == nil {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilSystemSCConfig)
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilAddressPubKeyConverter)
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilShardCoordinator)
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilEnableEpochsHandler)
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilNodesCoordinator)
	}

	scf := &systemSCFactory{
		systemEI:               args.SystemEI,
		sigVerifier:            args.SigVerifier,
		nodesConfigProvider:    args.NodesConfigProvider,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		systemSCConfig:         args.SystemSCConfig,
		economics:              args.Economics,
		addressPubKeyConverter: args.AddressPubKeyConverter,
		shardCoordinator:       args.ShardCoordinator,
		enableEpochsHandler:    args.EnableEpochsHandler,
		nodesCoordinator:       args.NodesCoordinator,
	}

	err := scf.createGasConfig(args.GasSchedule.LatestGasSchedule())
	if err != nil {
		return nil, err
	}

	scf.systemSCsContainer = NewSystemSCContainer()
	args.GasSchedule.RegisterNotifyHandler(scf)

	return scf, nil
}

func (scf *systemSCFactory) createGasConfig(gasMap map[string]map[string]uint64) error {
	baseOps := &vm.BaseOperationCost{}
	err := mapstructure.Decode(gasMap[common.BaseOperationCost], baseOps)
	if err != nil {
		return err
	}

	err = check.ForZeroUintFields(*baseOps)
	if err != nil {
		return err
	}

	metaChainSCsOps := &vm.MetaChainSystemSCsCost{}
	err = mapstructure.Decode(gasMap[common.MetaChainSystemSCsCost], metaChainSCsOps)
	if err != nil {
		return err
	}

	err = check.ForZeroUintFields(*metaChainSCsOps)
	if err != nil {
		return err
	}

	builtInFunctionsCost := &vm.BuiltInCost{}
	err = mapstructure.Decode(gasMap[common.BuiltInCost], builtInFunctionsCost)
	if err != nil {
		return err
	}

	scf.gasCost = vm.GasCost{
		BaseOperationCost:      *baseOps,
		MetaChainSystemSCsCost: *metaChainSCsOps,
		BuiltInCost:            *builtInFunctionsCost,
	}

	return nil
}

// GasScheduleChange is called when gas schedule is changed, thus all contracts must be updated
func (scf *systemSCFactory) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	err := scf.createGasConfig(gasSchedule)
	if err != nil {
		log.Error("error changing gas schedule", "error", err)
		return
	}

	var systemSC vm.SystemSmartContract
	for _, key := range scf.systemSCsContainer.Keys() {
		systemSC, err = scf.systemSCsContainer.Get(key)
		if err != nil {
			log.Error("error getting system SC", "key", key, "error", err)
			return
		}

		systemSC.SetNewGasCost(scf.gasCost)
	}

	log.Debug("new gas schedule was set")
}

func (scf *systemSCFactory) createStakingContract() (vm.SystemSmartContract, error) {
	argsStaking := systemSmartContracts.ArgsNewStakingSmartContract{
		MinNumNodes:          uint64(scf.nodesConfigProvider.MinNumberOfNodes()),
		StakingSCConfig:      scf.systemSCConfig.StakingSystemSCConfig,
		Eei:                  scf.systemEI,
		StakingAccessAddr:    vm.ValidatorSCAddress,
		JailAccessAddr:       vm.JailingAddress,
		EndOfEpochAccessAddr: vm.EndOfEpochAddress,
		GasCost:              scf.gasCost,
		Marshalizer:          scf.marshalizer,
		EnableEpochsHandler:  scf.enableEpochsHandler,
	}
	staking, err := systemSmartContracts.NewStakingSmartContract(argsStaking)
	return staking, err
}

func (scf *systemSCFactory) createValidatorContract() (vm.SystemSmartContract, error) {
	args := systemSmartContracts.ArgsValidatorSmartContract{
		Eei:                    scf.systemEI,
		SigVerifier:            scf.sigVerifier,
		StakingSCConfig:        scf.systemSCConfig.StakingSystemSCConfig,
		StakingSCAddress:       vm.StakingSCAddress,
		EndOfEpochAddress:      vm.EndOfEpochAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		GasCost:                scf.gasCost,
		Marshalizer:            scf.marshalizer,
		GenesisTotalSupply:     scf.economics.GenesisTotalSupply(),
		MinDeposit:             scf.systemSCConfig.DelegationManagerSystemSCConfig.MinCreationDeposit,
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		GovernanceSCAddress:    vm.GovernanceSCAddress,
		ShardCoordinator:       scf.shardCoordinator,
		EnableEpochsHandler:    scf.enableEpochsHandler,
		NodesCoordinator:       scf.nodesCoordinator,
	}
	validatorSC, err := systemSmartContracts.NewValidatorSmartContract(args)
	return validatorSC, err
}

func (scf *systemSCFactory) createESDTContract() (vm.SystemSmartContract, error) {
	argsESDT := systemSmartContracts.ArgsNewESDTSmartContract{
		Eei:                    scf.systemEI,
		GasCost:                scf.gasCost,
		ESDTSCAddress:          vm.ESDTSCAddress,
		Marshalizer:            scf.marshalizer,
		Hasher:                 scf.hasher,
		ESDTSCConfig:           scf.systemSCConfig.ESDTSystemSCConfig,
		AddressPubKeyConverter: scf.addressPubKeyConverter,
		EndOfEpochSCAddress:    vm.EndOfEpochAddress,
		EnableEpochsHandler:    scf.enableEpochsHandler,
		ESDTPrefix:             scf.systemSCConfig.ESDTSystemSCConfig.ESDTPrefix,
	}
	esdt, err := systemSmartContracts.NewESDTSmartContract(argsESDT)
	return esdt, err
}

func (scf *systemSCFactory) createGovernanceContract() (vm.SystemSmartContract, error) {
	ownerAddress, err := scf.addressPubKeyConverter.Decode(scf.systemSCConfig.GovernanceSystemSCConfig.OwnerAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for GovernanceSystemSCConfig.OwnerAddress in systemSCFactory", vm.ErrInvalidAddress)
	}

	argsGovernance := systemSmartContracts.ArgsNewGovernanceContract{
		Eei:                    scf.systemEI,
		GasCost:                scf.gasCost,
		GovernanceConfig:       scf.systemSCConfig.GovernanceSystemSCConfig,
		Marshalizer:            scf.marshalizer,
		Hasher:                 scf.hasher,
		GovernanceSCAddress:    vm.GovernanceSCAddress,
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		EnableEpochsHandler:    scf.enableEpochsHandler,
		UnBondPeriodInEpochs:   scf.systemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs,
		OwnerAddress:           ownerAddress,
	}
	governance, err := systemSmartContracts.NewGovernanceContract(argsGovernance)
	return governance, err
}

func (scf *systemSCFactory) createDelegationContract() (vm.SystemSmartContract, error) {
	addTokensAddress, err := scf.addressPubKeyConverter.Decode(scf.systemSCConfig.DelegationManagerSystemSCConfig.ConfigChangeAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for DelegationManagerSystemSCConfig.ConfigChangeAddress in systemSCFactory", vm.ErrInvalidAddress)
	}

	argsDelegation := systemSmartContracts.ArgsNewDelegation{
		DelegationSCConfig:     scf.systemSCConfig.DelegationSystemSCConfig,
		StakingSCConfig:        scf.systemSCConfig.StakingSystemSCConfig,
		Eei:                    scf.systemEI,
		SigVerifier:            scf.sigVerifier,
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		StakingSCAddress:       vm.StakingSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		GasCost:                scf.gasCost,
		Marshalizer:            scf.marshalizer,
		EndOfEpochAddress:      vm.EndOfEpochAddress,
		GovernanceSCAddress:    vm.GovernanceSCAddress,
		AddTokensAddress:       addTokensAddress,
		EnableEpochsHandler:    scf.enableEpochsHandler,
	}
	delegation, err := systemSmartContracts.NewDelegationSystemSC(argsDelegation)
	return delegation, err
}

func (scf *systemSCFactory) createDelegationManagerContract() (vm.SystemSmartContract, error) {
	configChangeAddres, err := scf.addressPubKeyConverter.Decode(scf.systemSCConfig.DelegationManagerSystemSCConfig.ConfigChangeAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for DelegationManagerSystemSCConfig.ConfigChangeAddress in systemSCFactory", vm.ErrInvalidAddress)
	}

	argsDelegationManager := systemSmartContracts.ArgsNewDelegationManager{
		DelegationMgrSCConfig:  scf.systemSCConfig.DelegationManagerSystemSCConfig,
		DelegationSCConfig:     scf.systemSCConfig.DelegationSystemSCConfig,
		Eei:                    scf.systemEI,
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		StakingSCAddress:       vm.StakingSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		ConfigChangeAddress:    configChangeAddres,
		GasCost:                scf.gasCost,
		Marshalizer:            scf.marshalizer,
		EnableEpochsHandler:    scf.enableEpochsHandler,
	}
	delegationManager, err := systemSmartContracts.NewDelegationManagerSystemSC(argsDelegationManager)
	return delegationManager, err
}

// CreateForGenesis instantiates all the system smart contracts and returns a container containing them to be used in the genesis process
func (scf *systemSCFactory) CreateForGenesis() (vm.SystemSCContainer, error) {
	staking, err := scf.createStakingContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.StakingSCAddress, staking)
	if err != nil {
		return nil, err
	}

	validatorSC, err := scf.createValidatorContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.ValidatorSCAddress, validatorSC)
	if err != nil {
		return nil, err
	}

	esdt, err := scf.createESDTContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.ESDTSCAddress, esdt)
	if err != nil {
		return nil, err
	}

	governance, err := scf.createGovernanceContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.GovernanceSCAddress, governance)
	if err != nil {
		return nil, err
	}

	err = scf.systemEI.SetSystemSCContainer(scf.systemSCsContainer)
	if err != nil {
		return nil, err
	}

	return scf.systemSCsContainer, nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	_, err := scf.CreateForGenesis()
	if err != nil {
		return nil, err
	}

	delegationManager, err := scf.createDelegationManagerContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.DelegationManagerSCAddress, delegationManager)
	if err != nil {
		return nil, err
	}

	delegation, err := scf.createDelegationContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.FirstDelegationSCAddress, delegation)
	if err != nil {
		return nil, err
	}

	err = scf.systemEI.SetSystemSCContainer(scf.systemSCsContainer)
	if err != nil {
		return nil, err
	}

	return scf.systemSCsContainer, nil
}

// IsInterfaceNil checks whether the underlying object is nil
func (scf *systemSCFactory) IsInterfaceNil() bool {
	return scf == nil
}
