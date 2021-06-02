package smartContract

type scProcessorSimulate struct {
	*scProcessor
}

// NewSmartContractProcessorSimulate will create a new instance of smart contract processor for simulate
func NewSmartContractProcessorSimulate(args ArgsNewSmartContractProcessor) (*scProcessorSimulate, error) {
	scProc, err := NewSmartContractProcessor(args)
	if err != nil {
		return nil, err
	}

	scProc.shouldCheckValues = func() bool {
		return false
	}

	return &scProcessorSimulate{
		scProcessor: scProc,
	}, nil
}
