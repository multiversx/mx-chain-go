package api

type sovereignAPIProcessorCompsCreator struct {
}

// NewSovereignAPIProcessorCompsCreator creates a new sovereign api processor components creator
func NewSovereignAPIProcessorCompsCreator() *sovereignAPIProcessorCompsCreator {
	return &sovereignAPIProcessorCompsCreator{}
}

// CreateAPIComps creates api process comps for sovereign, same as the ones from metachain
func (c *sovereignAPIProcessorCompsCreator) CreateAPIComps(args ArgsCreateAPIProcessComps) (*APIProcessComps, error) {
	return createAPIComps(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (c *sovereignAPIProcessorCompsCreator) IsInterfaceNil() bool {
	return c == nil
}
