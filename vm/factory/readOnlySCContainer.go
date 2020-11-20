package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/vm"
)

// readOnlySCContainer is a wrapper over a sccontainer db which works read-only. write operation are disabled
type readOnlySCContainer struct {
	scContainer vm.SystemSCContainer
}

// NewReadOnlySCContainer returns a new instance of readOnlySCContainer
func NewReadOnlySCContainer(scContainer vm.SystemSCContainer) (*readOnlySCContainer, error) {
	if check.IfNil(scContainer) {
		return nil, vm.ErrNilSystemContractsContainer
	}

	return &readOnlySCContainer{scContainer: scContainer}, nil
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (r *readOnlySCContainer) Get(key []byte) (vm.SystemSmartContract, error) {
	return r.scContainer.Get(key)
}

// Add is disabled from readOnlySCContainer
func (r *readOnlySCContainer) Add(_ []byte, _ vm.SystemSmartContract) error {
	return nil
}

// Replace is disabled from readOnlySCContainer
func (r *readOnlySCContainer) Replace(_ []byte, _ vm.SystemSmartContract) error {
	return nil
}

// Remove is disabled from readOnlySCContainer
func (r *readOnlySCContainer) Remove(_ []byte) {
}

// Len returns the number of containers
func (r *readOnlySCContainer) Len() int {
	return r.scContainer.Len()
}

// Keys returns the all the keys from the containers
func (r *readOnlySCContainer) Keys() [][]byte {
	return r.scContainer.Keys()
}

// IsInterfaceNil returns true if underlying object is nil
func (r *readOnlySCContainer) IsInterfaceNil() bool {
	return r == nil
}
