package mock

import "github.com/ElrondNetwork/elrond-go/data"

// GetHdrHandlerStub -
type GetHdrHandlerStub struct {
	HeaderHandlerCalled func() data.HeaderHandler
}

// HeaderHandler -
func (ghhs *GetHdrHandlerStub) HeaderHandler() data.HeaderHandler {
	return ghhs.HeaderHandlerCalled()
}
