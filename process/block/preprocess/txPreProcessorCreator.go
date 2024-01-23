package preprocess

import "github.com/multiversx/mx-chain-go/process"

type txPreProcessorCreator struct {
}

// NewTxPreProcessorCreator creates a tx pre-processor creator
func NewTxPreProcessorCreator() *txPreProcessorCreator {
	return &txPreProcessorCreator{}
}

// CreateTxPreProcessor creates a tx pre-processor for regular chain(meta+shard)
func (tpc *txPreProcessorCreator) CreateTxPreProcessor(args ArgsTransactionPreProcessor) (process.PreProcessor, error) {
	return NewTransactionPreprocessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (tpc *txPreProcessorCreator) IsInterfaceNil() bool {
	return tpc == nil
}
