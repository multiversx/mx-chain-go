package mock

import "github.com/ElrondNetwork/elrond-go-core/data"

// GetHdrHandlerStub -
type GetHdrHandlerStub struct {
	HeaderHandlerCalled func() data.HeaderHandler
}

// HeaderHandler -
func (ghhs *GetHdrHandlerStub) HeaderHandler() data.HeaderHandler {
	return ghhs.HeaderHandlerCalled()
}
