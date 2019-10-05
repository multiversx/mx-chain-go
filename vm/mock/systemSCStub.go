package mock

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type SystemSCStub struct {
	ExecuteCalled func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode
	ValueOfCalled func(key interface{}) interface{}
}

func (s *SystemSCStub) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.ExecuteCalled != nil {
		return s.ExecuteCalled(args)
	}
	return 0
}

func (s *SystemSCStub) ValueOf(key interface{}) interface{} {
	if s.ValueOfCalled != nil {
		return s.ValueOfCalled(key)
	}
	return nil
}

func (s *SystemSCStub) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
