package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HeaderValidatorStub struct {
	IsHeaderValidForProcessingCalled func(headerHandler data.HeaderHandler) bool
}

func (h *HeaderValidatorStub) IsHeaderValidForProcessing(headerHandler data.HeaderHandler) bool {
	return h.IsHeaderValidForProcessingCalled(headerHandler)
}
