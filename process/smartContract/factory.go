package smartContract

import (
	"github.com/multiversx/mx-chain-go/process"
)

type SCRProcessorFactory struct {
}

func (s *SCRProcessorFactory) CreateSCRProcessor(args process.ArgsNewSmartContractProcessor) (process.SCRProcessorHandler, error) {
	return NewSmartContractProcessor(args)
}
