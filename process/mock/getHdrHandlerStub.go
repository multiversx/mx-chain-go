package mock

import "github.com/ElrondNetwork/elrond-go/data"

type GetHdrHandlerStub struct {
	HeaderHandlerCalled func() data.HeaderHandler
}

func (ghhs *GetHdrHandlerStub) HeaderHandler() data.HeaderHandler {
	return ghhs.HeaderHandlerCalled()
}
