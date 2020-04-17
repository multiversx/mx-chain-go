package process

// ArgsHardForkBlockCreation
type ArgsHardForkBlockCreation struct {
}

type blockCreation struct {
}

// NewHardForkGenesisProcess
func NewHardForkGenesisProcess(args ArgsHardForkBlockCreation) (*blockCreation, error) {
	return &blockCreation{}, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (b *blockCreation) IsInterfaceNil() bool {
	return b == nil
}
