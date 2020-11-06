package mock

import (
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SystemSCStub -
type SystemSCStub struct {
	ExecuteCalled        func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode
	SetNewGasCostsCalled func(gasCost vm.GasCost)
}

// SetNewGasCosts -
func (s *SystemSCStub) SetNewGasCosts(gasCost vm.GasCost) {
	if s.SetNewGasCostsCalled != nil {
		s.SetNewGasCostsCalled(gasCost)
	}
}

// Execute -
func (s *SystemSCStub) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.ExecuteCalled != nil {
		return s.ExecuteCalled(args)
	}
	return 0
}

// IsInterfaceNil -
func (s *SystemSCStub) IsInterfaceNil() bool {
	return s == nil
}
