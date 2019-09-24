package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HdrValidatorHandlerStub struct {
	HashCalled          func() []byte
	HeaderHandlerCalled func() data.HeaderHandler
}

func (hvhs *HdrValidatorHandlerStub) Hash() []byte {
	return hvhs.HashCalled()
}

func (hvhs *HdrValidatorHandlerStub) HeaderHandler() data.HeaderHandler {
	return hvhs.HeaderHandlerCalled()
}
