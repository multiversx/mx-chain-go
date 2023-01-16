package mock

import "github.com/multiversx/mx-chain-core-go/data"

// GetHdrHandlerStub -
type GetHdrHandlerStub struct {
	HeaderHandlerCalled func() data.HeaderHandler
}

// HeaderHandler -
func (ghhs *GetHdrHandlerStub) HeaderHandler() data.HeaderHandler {
	return ghhs.HeaderHandlerCalled()
}
