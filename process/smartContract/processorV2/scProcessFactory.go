package processorV2

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
)

type scProcessFactory struct {
}

// NewSCProcessFactory creates a new smart contract process factory with proxy sc processor
func NewSCProcessFactory() (*scProcessFactory, error) {
	return &scProcessFactory{}, nil
}

// CreateSCProcessor creates a new smart contract processor
func (scpf *scProcessFactory) CreateSCProcessor(args scrCommon.ArgsNewSmartContractProcessor, _ process.EpochNotifier) (scrCommon.SCRProcessorHandler, error) {
	return NewSmartContractProcessorV2(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scpf *scProcessFactory) IsInterfaceNil() bool {
	return scpf == nil
}
