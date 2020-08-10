package mock

import "github.com/ElrondNetwork/elrond-go/genesis"

// DeployProcessorStub -
type DeployProcessorStub struct {
	DeployCalled                 func(sc genesis.InitialSmartContractHandler) ([][]byte, error)
	SetReplacePlaceholdersCalled func(handler func(txData string, scResultingAddressBytes []byte) (string, error))
}

// Deploy -
func (dps *DeployProcessorStub) Deploy(sc genesis.InitialSmartContractHandler) ([][]byte, error) {
	if dps.DeployCalled != nil {
		return dps.DeployCalled(sc)
	}

	return make([][]byte, 0), nil
}

// IsInterfaceNil -
func (dps *DeployProcessorStub) IsInterfaceNil() bool {
	return dps == nil
}
