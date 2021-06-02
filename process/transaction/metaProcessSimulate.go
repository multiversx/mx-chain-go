package transaction

type metaProcessorSimulate struct {
	*metaTxProcessor
}

// NewMetaTxProcessorSimulate will create a new instance of meta processor for simulate
func NewMetaTxProcessorSimulate(args ArgsNewMetaTxProcessor) (*metaProcessorSimulate, error) {
	txProc, err := NewMetaTxProcessor(args)
	if err != nil {
		return nil, err
	}

	txProc.shouldCheckBalanceHandler = func() bool {
		return false
	}

	return &metaProcessorSimulate{
		metaTxProcessor: txProc,
	}, nil
}
