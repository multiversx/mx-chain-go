package api

type sovereignAPIProcessorCompsCreator struct {
}

func NewSovereignAPIProcessorCompsCreator() *sovereignAPIProcessorCompsCreator {
	return &sovereignAPIProcessorCompsCreator{}
}

func (c *sovereignAPIProcessorCompsCreator) CreateAPIComps(args ArgsCreateAPIProcessComps) (*APIProcessComps, error) {
	return createAPIComps(args)
}

func (c *sovereignAPIProcessorCompsCreator) IsInterfaceNil() bool {
	return c == nil
}
