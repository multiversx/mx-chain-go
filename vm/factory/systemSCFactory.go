package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/mitchellh/mapstructure"
)

type systemSCFactory struct {
	systemEI          vm.ContextHandler
	validatorSettings vm.ValidatorSettingsHandler
	sigVerifier       vm.MessageSignVerifier
	gasCost           vm.GasCost
}

// ArgsNewSystemSCFactory defines the arguments struct needed to create the system SCs
type ArgsNewSystemSCFactory struct {
	SystemEI          vm.ContextHandler
	ValidatorSettings vm.ValidatorSettingsHandler
	SigVerifier       vm.MessageSignVerifier
	GasMap            map[string]map[string]uint64
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

	scf := &systemSCFactory{
		systemEI:          args.SystemEI,
		validatorSettings: args.ValidatorSettings,
		sigVerifier:       args.SigVerifier}

	err := scf.createGasConfig(args.GasMap)
	if err != nil {
		return nil, err
	}

	return scf, nil
}

func (scf *systemSCFactory) createGasConfig(gasMap map[string]map[string]uint64) error {
	baseOps := &vm.BaseOperationCost{}
	err := mapstructure.Decode(gasMap["BaseOperationCost"], baseOps)
	if err != nil {
		return err
	}

	err = core.CheckForZeroUint64Fields(*baseOps)
	if err != nil {
		return err
	}

	metaChainSCsOps := &vm.MetaChainSystemSCsCost{}
	err = mapstructure.Decode(gasMap["MetaChainSystemSCsCost"], metaChainSCsOps)
	if err != nil {
		return err
	}

	err = core.CheckForZeroUint64Fields(*metaChainSCsOps)
	if err != nil {
		return err
	}

	scf.gasCost = vm.GasCost{
		BaseOperationCost:      *baseOps,
		MetaChainSystemSCsCost: *metaChainSCsOps,
	}

	return nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	scContainer := NewSystemSCContainer()

	argsStaking := systemSmartContracts.ArgsNewStakingSmartContract{
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
		Eei:               scf.systemEI,
		SigVerifier:       scf.sigVerifier,
		ValidatorSettings: scf.validatorSettings,
		StakingSCAddress:  StakingSCAddress,
		AuctionSCAddress:  AuctionSCAddress,
		GasCost:           scf.gasCost,
	}
	auction, err := systemSmartContracts.NewStakingAuctionSmartContract(args)
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(AuctionSCAddress, auction)
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
