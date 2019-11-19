package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type SCQueryServiceStub struct {
	ExecuteQueryCalled func(*process.SCQuery) (*vmcommon.VMOutput, error)
}

func (serviceStub *SCQueryServiceStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return serviceStub.ExecuteQueryCalled(query)
}

// IsInterfaceNil returns true if there is no value under the interface
func (serviceStub *SCQueryServiceStub) IsInterfaceNil() bool {
	if serviceStub == nil {
		return true
	}
	return false
}
