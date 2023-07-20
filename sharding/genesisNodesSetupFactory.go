package sharding

type genesisNodesSetupFactory struct {
}

func NewGenesisNodesSetupFactory() GenesisNodesSetupFactory {
	return &genesisNodesSetupFactory{}
}

func (gns *genesisNodesSetupFactory) CreateNodesSetup(args *NodesSetupArgs) (GenesisNodesSetupHandler, error) {
	return NewNodesSetup(
		args.NodesFilePath,
		args.AddressPubKeyConverter,
		args.ValidatorPubKeyConverter,
		args.GenesisMaxNumShards,
	)
}

func (gns *genesisNodesSetupFactory) IsInterfaceNil() bool {
	return gns == nil
}
