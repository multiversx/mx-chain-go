package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

type systemSCFactory struct {
	systemEI          vm.ContextHandler
	validatorSettings vm.ValidatorSettingsHandler
	sigVerifier       vm.MessageSignVerifier
}

// NewSystemSCFactory creates a factory which will instantiate the system smart contracts
func NewSystemSCFactory(
	systemEI vm.ContextHandler,
	validatorSettings vm.ValidatorSettingsHandler,
	sigVerifier vm.MessageSignVerifier,
) (*systemSCFactory, error) {
	if check.IfNil(systemEI) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(validatorSettings) {
		return nil, vm.ErrNilEconomicsData
	}
	if check.IfNil(sigVerifier) {
		return nil, vm.ErrNilMessageSignVerifier
	}

	return &systemSCFactory{
		systemEI:          systemEI,
		validatorSettings: validatorSettings,
		sigVerifier:       sigVerifier}, nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	scContainer := NewSystemSCContainer()

	staking, err := systemSmartContracts.NewStakingSmartContract(
		scf.validatorSettings.StakeValue(),
		scf.validatorSettings.UnBondPeriod(),
		scf.systemEI,
		AuctionSCAddress,
	)
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
