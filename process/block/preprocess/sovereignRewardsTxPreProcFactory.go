package preprocess

import (
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignRewardsTxPreProcFactory struct {
}

// NewSovereignRewardsTxPreProcFactory creates a rewards tx pre-processor factory for sovereign chain
func NewSovereignRewardsTxPreProcFactory() *sovereignRewardsTxPreProcFactory {
	return &sovereignRewardsTxPreProcFactory{}
}

// CreateRewardsTxPreProcessorAndAddToContainer does not create any rewards processor, nor adds it to container, since they are not needed
func (f *sovereignRewardsTxPreProcFactory) CreateRewardsTxPreProcessorAndAddToContainer(_ ArgsRewardTxPreProcessor, _ process.PreProcessorsContainer) error {
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignRewardsTxPreProcFactory) IsInterfaceNil() bool {
	return f == nil
}
