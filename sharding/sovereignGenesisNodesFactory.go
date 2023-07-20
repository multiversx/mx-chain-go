package sharding

type sovereignGenesisNodesSetupFactory struct {
}

func NewSovereignGenesisNodesSetupFactory() GenesisNodesSetupFactory {
	return &sovereignGenesisNodesSetupFactory{}
}

func (gns *sovereignGenesisNodesSetupFactory) CreateNodesSetup(args *NodesSetupArgs) (GenesisNodesSetupHandler, error) {
	return NewSovereignNodesSetup(&SovereignNodesSetupArgs{
		NodesFilePath:            args.NodesFilePath,
		AddressPubKeyConverter:   args.AddressPubKeyConverter,
		ValidatorPubKeyConverter: args.ValidatorPubKeyConverter,
	})
}

func (gns *sovereignGenesisNodesSetupFactory) IsInterfaceNil() bool {
	return gns == nil
}
