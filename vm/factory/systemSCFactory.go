package factory

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

var initialStakeValue = "500000000000000000000000"

type systemSCFactory struct {
	systemEI vm.SystemEI
}

// NewSystemSCFactory creates a factory which will instantiate the system smart contracts
func NewSystemSCFactory(systemEI vm.SystemEI) (*systemSCFactory, error) {
	if systemEI == nil || systemEI.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	return &systemSCFactory{systemEI: systemEI}, nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	scContainer := NewSystemSCContainer()

	initValue, ok := big.NewInt(0).SetString(initialStakeValue, 10)
	if !ok {
		return nil, vm.ErrInvalidStakeValue
	}

	sc, err := systemSmartContracts.NewStakingSmartContract(initValue, scf.systemEI)
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
