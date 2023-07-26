package smartContract

type scProcessFactory struct {
}

// NewSCProcessFactory creates a new smart contract process factory
func NewSCProcessFactory() (*scProcessFactory, error) {
	return &scProcessFactory{}, nil
}

// CreateSCProcessor creates a new smart contract processor
func (scpf *scProcessFactory) CreateSCProcessor(args ArgsNewSmartContractProcessor) (SCRProcessorHandler, error) {
	return NewSmartContractProcessor(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scpf *scProcessFactory) IsInterfaceNil() bool {
	return scpf == nil
}
