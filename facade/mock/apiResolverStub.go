package mock

import "github.com/ElrondNetwork/elrond-go/statusHandler/nodeDetails"

type ApiResolverStub struct {
	GetVmValueHandler  func(address string, funcName string, argsBuff ...[]byte) ([]byte, error)
	NodeDetailsHandler func() nodeDetails.NodeDetails
}

func (ars *ApiResolverStub) GetVmValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return ars.GetVmValueHandler(address, funcName, argsBuff...)
}

func (ars *ApiResolverStub) NodeDetails() nodeDetails.NodeDetails {
	return ars.NodeDetailsHandler()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	if ars == nil {
		return true
	}
	return false
}
