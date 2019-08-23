package containers

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/cornelk/hashmap"
)

// interceptorsContainer is an interceptors holder organized by type
type interceptorsContainer struct {
	objects *hashmap.HashMap
}

// NewInterceptorsContainer will create a new instance of a container
func NewInterceptorsContainer() *interceptorsContainer {
	return &interceptorsContainer{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (ic *interceptorsContainer) Get(key string) (process.Interceptor, error) {
	value, ok := ic.objects.Get(key)
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}

	interceptor, ok := value.(process.Interceptor)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return interceptor, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (ic *interceptorsContainer) Add(key string, interceptor process.Interceptor) error {
	if interceptor == nil {
		return process.ErrNilContainerElement
	}

	ok := ic.objects.Insert(key, interceptor)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (ic *interceptorsContainer) AddMultiple(keys []string, interceptors []process.Interceptor) error {
	if len(keys) != len(interceptors) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := ic.Add(key, interceptors[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (ic *interceptorsContainer) Replace(key string, interceptor process.Interceptor) error {
	if interceptor == nil {
		return process.ErrNilContainerElement
	}

	ic.objects.Set(key, interceptor)
	return nil
}

// Remove will remove an object at a given key
func (ic *interceptorsContainer) Remove(key string) {
	ic.objects.Del(key)
}

// Len returns the length of the added objects
func (ic *interceptorsContainer) Len() int {
	return ic.objects.Len()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ic *interceptorsContainer) IsInterfaceNil() bool {
	if ic == nil {
		return true
	}
	return false
}
