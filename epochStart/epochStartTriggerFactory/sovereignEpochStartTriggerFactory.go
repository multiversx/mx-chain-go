package epochStartTriggerFactory

import (
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignEpochStartTriggerFactory struct {
}

// NewSovereignEpochStartTriggerFactory creates a sovereign epoch start trigger. This will be a metachain one, since
// nodes inside sovereign chain will not need to receive meta information, but they will actually execute the meta code.
func NewSovereignEpochStartTriggerFactory() *sovereignEpochStartTriggerFactory {
	return &sovereignEpochStartTriggerFactory{}
}

// CreateEpochStartTrigger creates a meta epoch start trigger for sovereign run type
func (f *sovereignEpochStartTriggerFactory) CreateEpochStartTrigger(args ArgsEpochStartTrigger) (process.EpochStartTriggerHandler, error) {
	return createMetaEpochStartTrigger(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignEpochStartTriggerFactory) IsInterfaceNil() bool {
	return f == nil
}
