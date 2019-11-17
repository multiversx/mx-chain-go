package factory

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

type systemSCFactory struct {
	systemEI          vm.SystemEI
	validatorSettings process.ValidatorSettingsHandler
}

// NewSystemSCFactory creates a factory which will instantiate the system smart contracts
func NewSystemSCFactory(
	systemEI vm.SystemEI,
	validatorSettings process.ValidatorSettingsHandler,
) (*systemSCFactory, error) {
	if systemEI == nil || systemEI.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if validatorSettings == nil || validatorSettings.IsInterfaceNil() {
		return nil, vm.ErrNilEconomicsData
	}

	return &systemSCFactory{
		systemEI:          systemEI,
		validatorSettings: validatorSettings}, nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	scContainer := NewSystemSCContainer()

	sc, err := systemSmartContracts.NewStakingSmartContract(
		scf.validatorSettings.StakeValue(),
		scf.validatorSettings.UnBoundPeriod(),
		scf.systemEI,
	)
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(StakingSCAddress, sc)
	if err != nil {
		return nil, err
	}

	return scContainer, nil
}

// IsInterfaceNil checks whether the underlying object is nil
func (scf *systemSCFactory) IsInterfaceNil() bool {
	if scf == nil {
		return true
	}
	return false
}
