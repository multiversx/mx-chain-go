package mock

import (
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type ApiResolverStub struct {
	ExecuteSCQueryHandler func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	StatusMetricsHandler  func() external.StatusMetricsHandler
}

func (ars *ApiResolverStub) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return ars.ExecuteSCQueryHandler(query)
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
