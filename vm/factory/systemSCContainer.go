package factory

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/cornelk/hashmap"
)

// systemSCContainer is an system smart contract holder organized by type
type systemSCContainer struct {
	objects *hashmap.HashMap
}

// NewSystemSCContainer will create a new instance of a container
func NewSystemSCContainer() *systemSCContainer {
	return &systemSCContainer{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (vmc *systemSCContainer) Get(key []byte) (vm.SystemSmartContract, error) {
	value, ok := vmc.objects.Get(key)
	if !ok {
		return nil, process.ErrInvalidContainerKey
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
	if sc == nil || sc.IsInterfaceNil() {
		return process.ErrNilContainerElement
	}

	ok := vmc.objects.Insert(key, sc)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (vmc *systemSCContainer) AddMultiple(keys [][]byte, scs []vm.SystemSmartContract) error {
	if len(keys) != len(scs) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := vmc.Add(key, scs[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (vmc *systemSCContainer) Replace(key []byte, sc vm.SystemSmartContract) error {
	if sc == nil || sc.IsInterfaceNil() {
		return process.ErrNilContainerElement
	}

	vmc.objects.Set(key, sc)
	return nil
}

// Remove will remove an object at a given key
func (vmc *systemSCContainer) Remove(key []byte) {
	vmc.objects.Del(key)
}

// Len returns the length of the added objects
func (vmc *systemSCContainer) Len() int {
	return vmc.objects.Len()
}

// Keys returns all the keys from the container
func (vmc *systemSCContainer) Keys() [][]byte {
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
func (vmc *systemSCContainer) IsInterfaceNil() bool {
	if vmc == nil {
		return true
	}
	return false
}
