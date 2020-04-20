package process

import "github.com/ElrondNetwork/elrond-go/data"

// ArgsAfterHardFork defines the arguments for the new after hard fork process handler
type ArgsAfterHardFork struct {
}

type afterHardFork struct {
}

// NewAfterHardForkBlockCreation creates the after hard fork block creator process handler
func NewAfterHardForkBlockCreation(args ArgsAfterHardFork) (*afterHardFork, error) {
	return &afterHardFork{}, nil
}

// CreateAllBlockAfterHardfork creates all the blocks after hardfork
func (a *afterHardFork) CreateAllBlockAfterHardfork() (map[uint32]data.HeaderHandler, map[uint32]data.BodyHandler, error) {
	return nil, nil, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (a *afterHardFork) IsInterfaceNil() bool {
	return a == nil
}
