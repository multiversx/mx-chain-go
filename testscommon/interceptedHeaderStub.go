package testscommon

import "github.com/ElrondNetwork/elrond-go-core/data"

// InterceptedHeaderStub -
type InterceptedHeaderStub struct {
	InterceptedDataStub
	HeaderHandlerCalled func() data.HeaderHandler
}

// HeaderHandler -
func (ihs *InterceptedHeaderStub) HeaderHandler() data.HeaderHandler {
	if ihs.HeaderHandlerCalled != nil {
		return ihs.HeaderHandlerCalled()
	}
	return nil
}
