package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/mitchellh/mapstructure"
)

type systemSCFactory struct {
	systemEI            vm.ContextHandler
	validatorSettings   vm.ValidatorSettingsHandler
	nodesConfigProvider vm.NodesConfigProvider
	sigVerifier         vm.MessageSignVerifier
	gasCost             vm.GasCost
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	systemSCConfig      *config.SystemSmartContractsConfig
}

// ArgsNewSystemSCFactory defines the arguments struct needed to create the system SCs
type ArgsNewSystemSCFactory struct {
	SystemEI            vm.ContextHandler
	ValidatorSettings   vm.ValidatorSettingsHandler
	NodesConfigProvider vm.NodesConfigProvider
	SigVerifier         vm.MessageSignVerifier
	GasMap              map[string]map[string]uint64
	Marshalizer         marshal.Marshalizer
	Hasher              hashing.Hasher
	SystemSCConfig      *config.SystemSmartContractsConfig
}

// NewSystemSCFactory creates a factory which will instantiate the system smart contracts
func NewSystemSCFactory(args ArgsNewSystemSCFactory) (*systemSCFactory, error) {
	if check.IfNil(args.SystemEI) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(args.ValidatorSettings) {
		return nil, vm.ErrNilEconomicsData
	}
	if check.IfNil(args.SigVerifier) {
		return nil, vm.ErrNilMessageSignVerifier
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, vm.ErrNilNodesConfigProvider
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, vm.ErrNilHasher
	}
	if args.SystemSCConfig == nil {
		return nil, vm.ErrNilSystemSCConfig
	}

	scf := &systemSCFactory{
		systemEI:            args.SystemEI,
		validatorSettings:   args.ValidatorSettings,
		sigVerifier:         args.SigVerifier,
		nodesConfigProvider: args.NodesConfigProvider,
		marshalizer:         args.Marshalizer,
		hasher:              args.Hasher,
		systemSCConfig:      args.SystemSCConfig,
	}

	err := scf.createGasConfig(args.GasMap)
	if err != nil {
		return nil, err
	}

	return scf, nil
}

func (scf *systemSCFactory) createGasConfig(gasMap map[string]map[string]uint64) error {
	baseOps := &vm.BaseOperationCost{}
	err := mapstructure.Decode(gasMap[core.BaseOperationCost], baseOps)
	if err != nil {
		return err
	}

	err = check.ForZeroUintFields(*baseOps)
	if err != nil {
		return err
	}

	metaChainSCsOps := &vm.MetaChainSystemSCsCost{}
	err = mapstructure.Decode(gasMap[core.MetaChainSystemSCsCost], metaChainSCsOps)
	if err != nil {
		return err
	}

	err = check.ForZeroUintFields(*metaChainSCsOps)
	if err != nil {
		return err
	}

	builtInFunctionsCost := &vm.BuiltInCost{}
	err = mapstructure.Decode(gasMap[core.BuiltInCost], builtInFunctionsCost)

	scf.gasCost = vm.GasCost{
		BaseOperationCost:      *baseOps,
		MetaChainSystemSCsCost: *metaChainSCsOps,
		BuiltInCost:            *builtInFunctionsCost,
	}

	return nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	scContainer := NewSystemSCContainer()

	argsStaking := systemSmartContracts.ArgsNewStakingSmartContract{
		MinNumNodes:              scf.nodesConfigProvider.MinNumberOfNodes(),
		MinStakeValue:            scf.validatorSettings.GenesisNodePrice(),
		UnBondPeriod:             scf.validatorSettings.UnBondPeriod(),
		Eei:                      scf.systemEI,
		StakingAccessAddr:        AuctionSCAddress,
		JailAccessAddr:           JailingAddress,
		NumRoundsWithoutBleed:    scf.validatorSettings.NumRoundsWithoutBleed(),
		BleedPercentagePerRound:  scf.validatorSettings.BleedPercentagePerRound(),
		MaximumPercentageToBleed: scf.validatorSettings.MaximumPercentageToBleed(),
		GasCost:                  scf.gasCost,
	}
	staking, err := systemSmartContracts.NewStakingSmartContract(argsStaking)
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(StakingSCAddress, staking)
	if err != nil {
		return nil, err
	}

	args := systemSmartContracts.ArgsStakingAuctionSmartContract{
		Eei:                 scf.systemEI,
		SigVerifier:         scf.sigVerifier,
		NodesConfigProvider: scf.nodesConfigProvider,
		ValidatorSettings:   scf.validatorSettings,
		StakingSCAddress:    StakingSCAddress,
		AuctionSCAddress:    AuctionSCAddress,
		GasCost:             scf.gasCost,
	}
	auction, err := systemSmartContracts.NewStakingAuctionSmartContract(args)
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(AuctionSCAddress, auction)
	if err != nil {
		return nil, err
	}

	argsESDT := systemSmartContracts.ArgsNewESDTSmartContract{
		Eei:           scf.systemEI,
		GasCost:       scf.gasCost,
		ESDTSCAddress: ESDTSCAddress,
		Marshalizer:   scf.marshalizer,
		Hasher:        scf.hasher,
		ESDTSCConfig:  scf.systemSCConfig.ESDTSystemSCConfig,
	}
	esdt, err := systemSmartContracts.NewESDTSmartContract(argsESDT)
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(ESDTSCAddress, esdt)
	if err != nil {
		return nil, err
	}

	err = scf.systemEI.SetSystemSCContainer(scContainer)
	if err != nil {
		return nil, err
	}

	return scContainer, nil
}

// IsInterfaceNil checks whether the underlying object is nil
func (scf *systemSCFactory) IsInterfaceNil() bool {
	return scf == nil
}
