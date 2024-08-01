package preprocess

import "github.com/multiversx/mx-chain-go/process"

type sovereignRewardsTxPreProcFactory struct {
}

// NewSovereignRewardsTxPreProcFactory creates a rewards tx pre-processor factory for sovereign chain
func NewSovereignRewardsTxPreProcFactory() *sovereignRewardsTxPreProcFactory {
	return &sovereignRewardsTxPreProcFactory{}
}

// CreateRewardsTxPreProcessor creates a rewards tx pre-processor for sovereign chain run type
func (f *sovereignRewardsTxPreProcFactory) CreateRewardsTxPreProcessor(args ArgsRewardTxPreProcessor) (process.PreProcessor, error) {
	return NewSovereignRewardsTxPreProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignRewardsTxPreProcFactory) IsInterfaceNil() bool {
	return f == nil
}
