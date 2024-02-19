package sharding

type genesisNodesSetupFactory struct {
}

// NewGenesisNodesSetupFactory creates a nodes setup factory for regular chain running(shards + metachain)
func NewGenesisNodesSetupFactory() GenesisNodesSetupFactory {
	return &genesisNodesSetupFactory{}
}

// CreateNodesSetup creates a genesis nodes setup handler for regular chain running(shards + metachain)
func (gns *genesisNodesSetupFactory) CreateNodesSetup(args *NodesSetupArgs) (GenesisNodesSetupHandler, error) {
	return NewNodesSetup(
		args.NodesFilePath,
		args.AddressPubKeyConverter,
		args.ValidatorPubKeyConverter,
		args.GenesisMaxNumShards,
	)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gns *genesisNodesSetupFactory) IsInterfaceNil() bool {
	return gns == nil
}
