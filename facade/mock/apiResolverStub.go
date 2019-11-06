package mock

import (
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type ApiResolverStub struct {
	SimulateRunSmartContractFunctionHandler func(command *smartContract.CommandRunFunction) (*vmcommon.VMOutput, error)
	StatusMetricsHandler                    func() external.StatusMetricsHandler
}

func (ars *ApiResolverStub) SimulateRunSmartContractFunction(command *smartContract.CommandRunFunction) (*vmcommon.VMOutput, error) {
	return ars.SimulateRunSmartContractFunctionHandler(command)
}

func (ars *ApiResolverStub) StatusMetrics() external.StatusMetricsHandler {
	return ars.StatusMetricsHandler()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	if ars == nil {
		return true
	}
	return false
}
