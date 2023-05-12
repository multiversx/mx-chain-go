package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type IncomingHeaderStub struct {
	GetHeaderHandlerCalled func() data.HeaderHandler
}

func (ihs *IncomingHeaderStub) GetIncomingEventHandlers() []data.EventHandler {
	return nil
}

func (ihs *IncomingHeaderStub) GetHeaderHandler() data.HeaderHandler {
	if ihs.GetHeaderHandlerCalled != nil {
		return ihs.GetHeaderHandlerCalled()
	}

	return &block.HeaderV2{}
}

func (ihs *IncomingHeaderStub) IsInterfaceNil() bool {
	return ihs == nil
}
