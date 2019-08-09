package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HeaderHandlerValidatorStub struct {
	CheckHeaderHandlerValidCalled func(headerHandler data.HeaderHandler) bool
}

func (h *HeaderHandlerValidatorStub) CheckHeaderHandlerValid(headerHandler data.HeaderHandler) bool {
	return h.CheckHeaderHandlerValidCalled(headerHandler)
}
