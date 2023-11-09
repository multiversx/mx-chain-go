package mock

import "github.com/multiversx/mx-chain-go/vm"

// SystemSCContainerStub -
type SystemSCContainerStub struct {
	GetCalled     func(key []byte) (vm.SystemSmartContract, error)
	AddCalled     func(key []byte, val vm.SystemSmartContract) error
	ReplaceCalled func(key []byte, val vm.SystemSmartContract) error
	RemoveCalled  func(key []byte)
	LenCalled     func() int
	KeysCalled    func() [][]byte
}

// Get -
func (s *SystemSCContainerStub) Get(key []byte) (vm.SystemSmartContract, error) {
	if s.GetCalled != nil {
		return s.GetCalled(key)
	}
	return nil, vm.ErrUnknownSystemSmartContract
}

// Add -
func (s *SystemSCContainerStub) Add(key []byte, val vm.SystemSmartContract) error {
	if s.AddCalled != nil {
		return s.AddCalled(key, val)
	}
	return nil
}

// Replace -
func (s *SystemSCContainerStub) Replace(key []byte, val vm.SystemSmartContract) error {
	if s.ReplaceCalled != nil {
		return s.ReplaceCalled(key, val)
	}
	return nil
}

// Remove -
func (s *SystemSCContainerStub) Remove(key []byte) {
	if s.RemoveCalled != nil {
		s.RemoveCalled(key)
	}
}

// Len -
func (s *SystemSCContainerStub) Len() int {
	if s.LenCalled != nil {
		return s.LenCalled()
	}
	return 0
}

// Keys -
func (s *SystemSCContainerStub) Keys() [][]byte {
	if s.KeysCalled != nil {
		return s.KeysCalled()
	}
	return nil
}

// IsInterfaceNil -
func (s *SystemSCContainerStub) IsInterfaceNil() bool {
	return s == nil
}
