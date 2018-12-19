package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type InterceptorStub struct {
	NameCalled                          func() string
	SetCheckReceivedObjectHandlerCalled func(func(newer p2p.Newer, rawData []byte) bool)
	CheckReceivedObjectHandlerCalled    func() func(newer p2p.Newer, rawData []byte) bool
}

func (is *InterceptorStub) Name() string {
	return is.NameCalled()
}

func (is *InterceptorStub) SetCheckReceivedObjectHandler(handler func(newer p2p.Newer, rawData []byte) bool) {
	is.SetCheckReceivedObjectHandlerCalled(handler)
}

func (is *InterceptorStub) CheckReceivedObjectHandler() func(newer p2p.Newer, rawData []byte) bool {
	return is.CheckReceivedObjectHandlerCalled()
}
