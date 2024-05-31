package process

type sovereignGenesisBlockCreatorFactory struct {
}

// NewSovereignGenesisBlockCreatorFactory creates a sovereign genesis block creator factory
func NewSovereignGenesisBlockCreatorFactory() *sovereignGenesisBlockCreatorFactory {
	return &sovereignGenesisBlockCreatorFactory{}
}

// CreateGenesisBlockCreator creates a sovereign genesis block creator
func (gbf *sovereignGenesisBlockCreatorFactory) CreateGenesisBlockCreator(args ArgsGenesisBlockCreator) (GenesisBlockCreatorHandler, error) {
	gbc, err := NewGenesisBlockCreator(args)
	if err != nil {
		return nil, nil
	}

	return NewSovereignGenesisBlockCreator(gbc)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gbf *sovereignGenesisBlockCreatorFactory) IsInterfaceNil() bool {
	return gbf == nil
}
