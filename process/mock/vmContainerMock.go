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
func (v *VMContainerMock) Get(key []byte) (vmcommon.VMExecutionHandler, error) {
	if v.GetCalled == nil {
		return &VMExecutionHandlerStub{}, nil
	}
	return v.GetCalled(key)
}

// Add -
func (v *VMContainerMock) Add(key []byte, val vmcommon.VMExecutionHandler) error {
	if v.AddCalled == nil {
		return nil
	}
	return v.AddCalled(key, val)
}

// AddMultiple -
func (v *VMContainerMock) AddMultiple(keys [][]byte, vms []vmcommon.VMExecutionHandler) error {
	if v.AddMultipleCalled == nil {
		return nil
	}
	return v.AddMultipleCalled(keys, vms)
}

// Replace -
func (v *VMContainerMock) Replace(key []byte, val vmcommon.VMExecutionHandler) error {
	if v.ReplaceCalled == nil {
		return nil
	}
	return v.ReplaceCalled(key, val)
}

// Remove -
func (v *VMContainerMock) Remove(key []byte) {
	if v.RemoveCalled == nil {
		return
	}
	v.RemoveCalled(key)
}

// Len -
func (v *VMContainerMock) Len() int {
	if v.LenCalled == nil {
		return 0
	}
	return v.LenCalled()
}

// Keys -
func (v *VMContainerMock) Keys() [][]byte {
	if v.KeysCalled == nil {
		return nil
	}
	return v.KeysCalled()
}

// Close -
func (v *VMContainerMock) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *VMContainerMock) IsInterfaceNil() bool {
	return v == nil
}
