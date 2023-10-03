package preprocess

import "github.com/multiversx/mx-chain-go/process"

type sovereignTxPreProcessorCreator struct {
}

// NewSovereignTxPreprocessorCreator creates a sovereign tx processor creator
func NewSovereignTxPreprocessorCreator() *sovereignTxPreProcessorCreator {
	return &sovereignTxPreProcessorCreator{}
}

// CreateTxProcessor creates a tx processor creator for sovereign chain
func (tpc *sovereignTxPreProcessorCreator) CreateTxProcessor(args ArgsTransactionPreProcessor) (process.PreProcessor, error) {
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
