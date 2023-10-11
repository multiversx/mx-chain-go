package mock

import (
	"errors"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var errNotImplemented = errors.New("not implemented")

// BuiltInFunctionContainerStub -
type BuiltInFunctionContainerStub struct {
}

// Get returns nil and error
func (bifc *BuiltInFunctionContainerStub) Get(_ string) (vmcommon.BuiltinFunction, error) {
	return nil, errNotImplemented
}

// Add does nothing and returns nil error
func (bifc *BuiltInFunctionContainerStub) Add(_ string, _ vmcommon.BuiltinFunction) error {
	return nil
}

// Replace does nothing and returns nil error
func (bifc *BuiltInFunctionContainerStub) Replace(_ string, _ vmcommon.BuiltinFunction) error {
	return nil
}

// Remove does nothing
func (bifc *BuiltInFunctionContainerStub) Remove(_ string) {
}

// Len returns 0
func (bifc *BuiltInFunctionContainerStub) Len() int {
	return 0
}

// Keys returns an empty map
func (bifc *BuiltInFunctionContainerStub) Keys() map[string]struct{} {
	return make(map[string]struct{})
}

// IsInterfaceNil returns true if there is no value under the interface
func (bifc *BuiltInFunctionContainerStub) IsInterfaceNil() bool {
	return bifc == nil
}
