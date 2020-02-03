package mock

import vmcommon "github.com/ElrondNetwork/elrond-vm-common"

type VMContainerMock struct {
	GetCalled         func(key []byte) (vmcommon.VMExecutionHandler, error)
	AddCalled         func(key []byte, val vmcommon.VMExecutionHandler) error
	AddMultipleCalled func(keys [][]byte, preprocessors []vmcommon.VMExecutionHandler) error
	ReplaceCalled     func(key []byte, val vmcommon.VMExecutionHandler) error
	RemoveCalled      func(key []byte)
	LenCalled         func() int
	KeysCalled        func() [][]byte
}

func (v *VMContainerMock) Get(key []byte) (vmcommon.VMExecutionHandler, error) {
	if v.GetCalled == nil {
		return &VMExecutionHandlerStub{}, nil
	}
	return v.GetCalled(key)
}

func (v *VMContainerMock) Add(key []byte, val vmcommon.VMExecutionHandler) error {
	if v.AddCalled == nil {
		return nil
	}
	return v.AddCalled(key, val)
}

func (v *VMContainerMock) AddMultiple(keys [][]byte, vms []vmcommon.VMExecutionHandler) error {
	if v.AddMultipleCalled == nil {
		return nil
	}
	return v.AddMultipleCalled(keys, vms)
}

func (v *VMContainerMock) Replace(key []byte, val vmcommon.VMExecutionHandler) error {
	if v.ReplaceCalled == nil {
		return nil
	}
	return v.ReplaceCalled(key, val)
}

func (v *VMContainerMock) Remove(key []byte) {
	if v.RemoveCalled == nil {
		return
	}
	v.RemoveCalled(key)
}

func (v *VMContainerMock) Len() int {
	if v.LenCalled == nil {
		return 0
	}
	return v.LenCalled()
}

func (v *VMContainerMock) Keys() [][]byte {
	if v.KeysCalled == nil {
		return nil
	}
	return v.KeysCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *VMContainerMock) IsInterfaceNil() bool {
	return v == nil
}
