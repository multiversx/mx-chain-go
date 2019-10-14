package mock

import "github.com/ElrondNetwork/elrond-go/vm"

type SystemSCContainerStub struct {
	GetCalled func(key []byte) (vm.SystemSmartContract, error)
	AddCalled func(key []byte, val vm.SystemSmartContract) error
	ReplaceCalled func(key []byte, val vm.SystemSmartContract) error
	RemoveCalled func(key []byte)
	LenCalled func() int
	KeysCalled func() [][]byte
}

func (s *SystemSCContainerStub) Get(key []byte) (vm.SystemSmartContract, error) {
	if s.GetCalled != nil {
		return s.GetCalled(key)
	}
	return nil, vm.ErrUnknownSystemSmartContract
}

func (s *SystemSCContainerStub) Add(key []byte, val vm.SystemSmartContract) error {
	if s.AddCalled != nil {
		return s.AddCalled(key, val)
	}
	return nil
}

func (s *SystemSCContainerStub) Replace(key []byte, val vm.SystemSmartContract) error {
	if s.ReplaceCalled != nil {
		return s.ReplaceCalled(key, val)
	}
	return nil
}

func (s *SystemSCContainerStub) Remove(key []byte) {
	if s.RemoveCalled != nil {
		s.RemoveCalled(key)
	}
	return
}

func (s *SystemSCContainerStub) Len() int {
	if s.LenCalled != nil {
		return s.LenCalled()
	}
	return 0
}

func (s *SystemSCContainerStub) Keys() [][]byte {
	if s.KeysCalled != nil {
		return s.KeysCalled()
	}
	return nil
}

func (s *SystemSCContainerStub) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}

