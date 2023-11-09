package factory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
)

// systemSCContainer is an system smart contract holder organized by type
type systemSCContainer struct {
	objects *container.MutexMap
}

// NewSystemSCContainer will create a new instance of a container
func NewSystemSCContainer() *systemSCContainer {
	return &systemSCContainer{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (vmc *systemSCContainer) Get(key []byte) (vm.SystemSmartContract, error) {
	value, ok := vmc.objects.Get(string(key))
	if !ok {
		return nil, fmt.Errorf("%w in system sc container for key %v", process.ErrInvalidContainerKey, key)
	}

	sc, ok := value.(vm.SystemSmartContract)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return sc, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (vmc *systemSCContainer) Add(key []byte, sc vm.SystemSmartContract) error {
	if check.IfNil(sc) {
		return process.ErrNilContainerElement
	}
	if len(key) == 0 {
		return vm.ErrNilOrEmptyKey
	}

	ok := vmc.objects.Insert(string(key), sc)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (vmc *systemSCContainer) Replace(key []byte, sc vm.SystemSmartContract) error {
	if check.IfNil(sc) {
		return process.ErrNilContainerElement
	}
	if len(key) == 0 {
		return vm.ErrNilOrEmptyKey
	}

	vmc.objects.Set(string(key), sc)
	return nil
}

// Remove will remove an object at a given key
func (vmc *systemSCContainer) Remove(key []byte) {
	vmc.objects.Remove(string(key))
}

// Len returns the length of the added objects
func (vmc *systemSCContainer) Len() int {
	return vmc.objects.Len()
}

// Keys returns all the keys from the container
func (vmc *systemSCContainer) Keys() [][]byte {
	keys := vmc.objects.Keys()
	keysBytes := make([][]byte, 0, len(keys))
	for _, k := range keys {
		key, ok := k.(string)
		if !ok {
			continue
		}
		keysBytes = append(keysBytes, []byte(key))
	}

	return keysBytes
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmc *systemSCContainer) IsInterfaceNil() bool {
	return vmc == nil
}
