package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type HeaderValidatorStub struct {
	HeaderValidForProcessingCalled func(headerHandler process.HdrValidatorHandler) error
}

func (h *HeaderValidatorStub) HeaderValidForProcessing(headerHandler process.HdrValidatorHandler) error {
	return h.HeaderValidForProcessingCalled(headerHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *HeaderValidatorStub) IsInterfaceNil() bool {
	if h == nil {
		return true
	}
	return false
}
