package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type InterceptorStub struct {
	NameCalled                          func() string
	SetCheckReceivedObjectHandlerCalled func(func(newer p2p.Creator, rawData []byte) error)
	CheckReceivedObjectHandlerCalled    func() func(newer p2p.Creator, rawData []byte) error
}

func (is *InterceptorStub) Name() string {
	return is.NameCalled()
}

func (is *InterceptorStub) SetCheckReceivedObjectHandler(handler func(newer p2p.Creator, rawData []byte) error) {
	is.SetCheckReceivedObjectHandlerCalled(handler)
}

func (is *InterceptorStub) CheckReceivedObjectHandler() func(newer p2p.Creator, rawData []byte) error {
	return is.CheckReceivedObjectHandlerCalled()
}
