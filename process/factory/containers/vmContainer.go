package containers

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/cornelk/hashmap"
)

// virtualMachinesContainer is an VM holder organized by type
type virtualMachinesContainer struct {
	objects *hashmap.HashMap
}

// NewVirtualMachinesContainer will create a new instance of a container
func NewVirtualMachinesContainer() *virtualMachinesContainer {
	return &virtualMachinesContainer{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (vmc *virtualMachinesContainer) Get(key []byte) (vmcommon.VMExecutionHandler, error) {
	value, ok := vmc.objects.Get(key)
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}

	vm, ok := value.(vmcommon.VMExecutionHandler)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return vm, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (vmc *virtualMachinesContainer) Add(key []byte, vm vmcommon.VMExecutionHandler) error {
	if vm == nil {
		return process.ErrNilContainerElement
	}

	ok := vmc.objects.Insert(key, vm)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (vmc *virtualMachinesContainer) AddMultiple(keys [][]byte, vms []vmcommon.VMExecutionHandler) error {
	if len(keys) != len(vms) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := vmc.Add(key, vms[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (vmc *virtualMachinesContainer) Replace(key []byte, vm vmcommon.VMExecutionHandler) error {
	if vm == nil {
		return process.ErrNilContainerElement
	}

	vmc.objects.Set(key, vm)
	return nil
}

// Remove will remove an object at a given key
func (vmc *virtualMachinesContainer) Remove(key []byte) {
	vmc.objects.Del(key)
}

// Len returns the length of the added objects
func (vmc *virtualMachinesContainer) Len() int {
	return vmc.objects.Len()
}

// Keys returns all the keys from the container
func (vmc *virtualMachinesContainer) Keys() [][]byte {
	keys := make([][]byte, 0)
	for obj := range vmc.objects.Iter() {
		byteKey, ok := obj.Key.([]byte)
		if !ok {
			continue
		}

		keys = append(keys, byteKey)
	}
	return keys
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmc *virtualMachinesContainer) IsInterfaceNil() bool {
	if vmc == nil {
		return true
	}
	return false
}
