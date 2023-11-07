package preprocess

import "github.com/multiversx/mx-chain-go/process"

type sovereignTxPreProcessorCreator struct {
}

// NewSovereignTxPreProcessorCreator creates a sovereign tx pre-processor creator
func NewSovereignTxPreProcessorCreator() *sovereignTxPreProcessorCreator {
	return &sovereignTxPreProcessorCreator{}
}

// CreateTxPreProcessor creates a tx pre-processor for sovereign chain
func (tpc *sovereignTxPreProcessorCreator) CreateTxPreProcessor(args ArgsTransactionPreProcessor) (process.PreProcessor, error) {
	txPreProc, err := NewTransactionPreprocessor(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignChainTransactionPreprocessor(txPreProc)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (tpc *sovereignTxPreProcessorCreator) IsInterfaceNil() bool {
	return tpc == nil
}
