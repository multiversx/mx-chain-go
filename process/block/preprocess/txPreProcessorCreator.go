package preprocess

import "github.com/multiversx/mx-chain-go/process"

type txPreProcessorCreator struct {
}

// NewTxPreProcessorCreator creates a tx processor creator
func NewTxPreProcessorCreator() *txPreProcessorCreator {
	return &txPreProcessorCreator{}
}

// CreateTxProcessor creates a tx processor creator for regular chain(meta+shard)
func (tpc *txPreProcessorCreator) CreateTxProcessor(args ArgsTransactionPreProcessor) (process.PreProcessor, error) {
	return NewTransactionPreprocessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (tpc *txPreProcessorCreator) IsInterfaceNil() bool {
	return tpc == nil
}
