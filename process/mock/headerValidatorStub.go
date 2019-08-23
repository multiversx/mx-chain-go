package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HeaderValidatorStub struct {
	IsHeaderValidForProcessingCalled func(headerHandler data.HeaderHandler) bool
}

func (h *HeaderValidatorStub) IsHeaderValidForProcessing(headerHandler data.HeaderHandler) bool {
	return h.IsHeaderValidForProcessingCalled(headerHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *HeaderValidatorStub) IsInterfaceNil() bool {
	if h == nil {
		return true
	}
	return false
}
