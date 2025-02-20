package sharding

type sovereignGenesisNodesSetupFactory struct {
}

// NewSovereignGenesisNodesSetupFactory creates a nodes setup factory for sovereign chain
func NewSovereignGenesisNodesSetupFactory() GenesisNodesSetupFactory {
	return &sovereignGenesisNodesSetupFactory{}
}

// CreateNodesSetup creates a genesis nodes setup handler for sovereign chain
func (gns *sovereignGenesisNodesSetupFactory) CreateNodesSetup(args *NodesSetupArgs) (GenesisNodesSetupHandler, error) {
	return NewSovereignNodesSetup(&SovereignNodesSetupArgs{
		NodesFilePath:            args.NodesFilePath,
		AddressPubKeyConverter:   args.AddressPubKeyConverter,
		ValidatorPubKeyConverter: args.ValidatorPubKeyConverter,
	})
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gns *sovereignGenesisNodesSetupFactory) IsInterfaceNil() bool {
	return gns == nil
}
