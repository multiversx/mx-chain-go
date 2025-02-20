package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignSmartContractResultPreProcessorFactory struct {
	smartContractResultPreProcessorCreator SmartContractResultPreProcessorCreator
}

// NewSovereignSmartContractResultPreProcessorFactory creates a new smart contract result pre processor factory for sovereign
func NewSovereignSmartContractResultPreProcessorFactory(creator SmartContractResultPreProcessorCreator) (*sovereignSmartContractResultPreProcessorFactory, error) {
	if check.IfNil(creator) {
		return nil, process.ErrNilPreProcessorCreator
	}

	return &sovereignSmartContractResultPreProcessorFactory{
		smartContractResultPreProcessorCreator: creator,
	}, nil
}

// CreateSmartContractResultPreProcessor creates a new smart contract result pre processor
func (scrppf *sovereignSmartContractResultPreProcessorFactory) CreateSmartContractResultPreProcessor(args SmartContractResultPreProcessorCreatorArgs) (process.PreProcessor, error) {
	scrpp, err := scrppf.smartContractResultPreProcessorCreator.CreateSmartContractResultPreProcessor(args)
	if err != nil {
		return nil, err
	}
	smartContractPreProcess, ok := scrpp.(*smartContractResults)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}
	return NewSovereignChainIncomingSCR(smartContractPreProcess)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scrppf *sovereignSmartContractResultPreProcessorFactory) IsInterfaceNil() bool {
	return scrppf == nil
}
