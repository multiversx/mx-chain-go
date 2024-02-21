package processProxy

import (
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
)

type scProcessProxyFactory struct {
}

// NewSCProcessProxyFactory creates a new smart contract process factory with proxy sc processor
func NewSCProcessProxyFactory() *scProcessProxyFactory {
	return &scProcessProxyFactory{}
}

// CreateSCProcessor creates a new smart contract processor
func (scpf *scProcessProxyFactory) CreateSCProcessor(args scrCommon.ArgsNewSmartContractProcessor) (scrCommon.SCRProcessorHandler, error) {
	return NewSmartContractProcessorProxy(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scpf *scProcessProxyFactory) IsInterfaceNil() bool {
	return scpf == nil
}
