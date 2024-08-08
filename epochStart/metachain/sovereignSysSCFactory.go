package metachain

import "github.com/multiversx/mx-chain-go/process"

type sovereignSysSCFactory struct {
}

// NewSovereignSysSCFactory creates a sys sc processor factory for sovereign chain
func NewSovereignSysSCFactory() *sovereignSysSCFactory {
	return &sovereignSysSCFactory{}
}

// CreateSystemSCProcessor creates a sys sc processor for sovereign chain
func (f *sovereignSysSCFactory) CreateSystemSCProcessor(args ArgsNewEpochStartSystemSCProcessing) (process.EpochStartSystemSCProcessor, error) {
	sysSC, err := NewSystemSCProcessor(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignSystemSCProcessor(sysSC)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignSysSCFactory) IsInterfaceNil() bool {
	return f == nil
}
