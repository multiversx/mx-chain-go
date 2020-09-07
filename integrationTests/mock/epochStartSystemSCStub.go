package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// EpochStartSystemSCStub -
type EpochStartSystemSCStub struct {
	ProcessSystemSmartContractCalled func(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error
}

// ProcessSystemSmartContract -
func (e *EpochStartSystemSCStub) ProcessSystemSmartContract(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error {
	if e.ProcessSystemSmartContractCalled != nil {
		return e.ProcessSystemSmartContractCalled(validatorInfos, epoch)
	}
	return nil
}

// IsInterfaceNil -
func (e *EpochStartSystemSCStub) IsInterfaceNil() bool {
	return e == nil
}
