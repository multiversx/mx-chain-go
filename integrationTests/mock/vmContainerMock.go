package mock

import vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"

// VMContainerMock -
type VMContainerMock struct {
	GetCalled         func(key []byte) (vmcommon.VMExecutionHandler, error)
	AddCalled         func(key []byte, val vmcommon.VMExecutionHandler) error
	AddMultipleCalled func(keys [][]byte, preprocessors []vmcommon.VMExecutionHandler) error
	ReplaceCalled     func(key []byte, val vmcommon.VMExecutionHandler) error
	RemoveCalled      func(key []byte)
	LenCalled         func() int
	KeysCalled        func() [][]byte
}

// Get -
func (vmc *VMContainerMock) Get(key []byte) (vmcommon.VMExecutionHandler, error) {
	if vmc.GetCalled == nil {
		return &VMExecutionHandlerStub{}, nil
	}
	return vmc.GetCalled(key)
}

// Add -
func (vmc *VMContainerMock) Add(key []byte, val vmcommon.VMExecutionHandler) error {
	if vmc.AddCalled == nil {
		return nil
	}
	return vmc.AddCalled(key, val)
}

// AddMultiple -
func (vmc *VMContainerMock) AddMultiple(keys [][]byte, vms []vmcommon.VMExecutionHandler) error {
	if vmc.AddMultipleCalled == nil {
		return nil
	}
	return vmc.AddMultipleCalled(keys, vms)
}

// Replace -
func (vmc *VMContainerMock) Replace(key []byte, val vmcommon.VMExecutionHandler) error {
	if vmc.ReplaceCalled == nil {
		return nil
	}
	return vmc.ReplaceCalled(key, val)
}

// Remove -
func (vmc *VMContainerMock) Remove(key []byte) {
	if vmc.RemoveCalled == nil {
		return
	}
	vmc.RemoveCalled(key)
}

// Len -
func (vmc *VMContainerMock) Len() int {
	if vmc.LenCalled == nil {
		return 0
	}
	return vmc.LenCalled()
}

// Keys -
func (vmc *VMContainerMock) Keys() [][]byte {
	if vmc.KeysCalled == nil {
		return nil
	}
	return vmc.KeysCalled()
}

// Close -
func (V *VMContainerMock) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmc *VMContainerMock) IsInterfaceNil() bool {
	return vmc == nil
}
