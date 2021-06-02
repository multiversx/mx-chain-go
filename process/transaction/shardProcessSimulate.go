package transaction

type shardProcessorSimulate struct {
	*txProcessor
}

// NewTxProcessorSimulate will create a new instance of shard processor for simulate
func NewTxProcessorSimulate(args ArgsNewTxProcessor) (*shardProcessorSimulate, error) {
	txProc, err := NewTxProcessor(args)
	if err != nil {
		return nil, err
	}

	txProc.shouldCheckBalanceHandler = func() bool {
		return false
	}

	return &shardProcessorSimulate{
		txProcessor: txProc,
	}, nil
}
