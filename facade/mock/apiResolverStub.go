package mock

import (
	"github.com/ElrondNetwork/elrond-go/node/external"
)

type ApiResolverStub struct {
	GetVmValueHandler    func(address string, funcName string, argsBuff ...[]byte) ([]byte, error)
	StatusMetricsHandler func() external.StatusMetricsHandler
}

func (ars *ApiResolverStub) GetVmValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return ars.GetVmValueHandler(address, funcName, argsBuff...)
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
