package mock

import (
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
)

type SCQueryServiceStub struct {
	ExecuteQueryCalled func(smartContract.SCQuery) ([]byte, error)
}

func (serviceStub *SCQueryServiceStub) ExecuteQuery(query smartContract.SCQuery) ([]byte, error) {
	return serviceStub.ExecuteQueryCalled(query)
}

// IsInterfaceNil returns true if there is no value under the interface
func (serviceStub *SCQueryServiceStub) IsInterfaceNil() bool {
	if serviceStub == nil {
		return true
	}
	return false
}
