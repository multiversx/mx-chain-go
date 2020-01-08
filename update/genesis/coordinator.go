package genesis

type ArgsNewGenesisCreator struct {
}

type genesisCreator struct {
}

func NewGenesisCreator() (*genesisCreator, error) {
	return nil, nil
}

func (g *genesisCreator) IsInterfaceNil() bool {
	return g == nil
}
