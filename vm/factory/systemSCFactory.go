package factory

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

type systemSCFactory struct {
	systemEI   vm.SystemEI
	nodesSetup *sharding.NodesSetup
}

// NewSystemSCFactory creates a factory which will instantiate the system smart contracts
func NewSystemSCFactory(
	systemEI vm.SystemEI,
	nodesSetup *sharding.NodesSetup,
) (*systemSCFactory, error) {
	if systemEI == nil || systemEI.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if nodesSetup == nil {
		return nil, vm.ErrNilNodesSetup
	}
	if nodesSetup.StakedValue.Cmp(big.NewInt(0)) < 0 {
		return nil, vm.ErrInvalidStakeValue
	}

	return &systemSCFactory{
		systemEI:   systemEI,
		nodesSetup: nodesSetup}, nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	scContainer := NewSystemSCContainer()

	sc, err := systemSmartContracts.NewStakingSmartContract(scf.nodesSetup.StakedValue, scf.systemEI)
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
