package process

type genesisBlockCreatorFactory struct {
}

// NewGenesisBlockCreatorFactory creates a genesis block creator factory
func NewGenesisBlockCreatorFactory() *genesisBlockCreatorFactory {
	return &genesisBlockCreatorFactory{}
}

// CreateGenesisBlockCreator creates a genesis block creator
func (gbf *genesisBlockCreatorFactory) CreateGenesisBlockCreator(args ArgsGenesisBlockCreator) (GenesisBlockCreatorHandler, error) {
	return NewGenesisBlockCreator(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gbf *genesisBlockCreatorFactory) IsInterfaceNil() bool {
	return gbf == nil
}
