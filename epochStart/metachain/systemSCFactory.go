package metachain

import "github.com/multiversx/mx-chain-go/process"

type sysSCFactory struct {
}

// NewSysSCFactory creates a sys sc processor factory for normal chain (meta processing)
func NewSysSCFactory() *sysSCFactory {
	return &sysSCFactory{}
}

// CreateSystemSCProcessor creates a sys sc processor for normal chain (meta processing)
func (f *sysSCFactory) CreateSystemSCProcessor(args ArgsNewEpochStartSystemSCProcessing) (process.EpochStartSystemSCProcessor, error) {
	return NewSystemSCProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sysSCFactory) IsInterfaceNil() bool {
	return f == nil
}
