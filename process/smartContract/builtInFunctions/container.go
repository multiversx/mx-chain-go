package builtInFunctions

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/container"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.BuiltInFunctionContainer = (*functionContainer)(nil)

// functionContainer is an interceptors holder organized by type
type functionContainer struct {
	objects *container.MutexMap
}

// NewBuiltInFunctionContainer will create a new instance of a container
func NewBuiltInFunctionContainer() *functionContainer {
	return &functionContainer{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (f *functionContainer) Get(key string) (process.BuiltinFunction, error) {
	value, ok := f.objects.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w in function container for key %v", process.ErrInvalidContainerKey, key)
	}

	function, ok := value.(process.BuiltinFunction)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return function, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (f *functionContainer) Add(key string, function process.BuiltinFunction) error {
	if check.IfNil(function) {
		return process.ErrNilContainerElement
	}
	if len(key) == 0 {
		return process.ErrEmptyFunctionName
	}

	ok := f.objects.Insert(key, function)
	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (f *functionContainer) Replace(key string, function process.BuiltinFunction) error {
	if check.IfNil(function) {
		return process.ErrNilContainerElement
	}
	if len(key) == 0 {
		return process.ErrEmptyFunctionName
	}

	f.objects.Set(key, function)
	return nil
}

// Remove will remove an object at a given key
func (f *functionContainer) Remove(key string) {
	f.objects.Remove(key)
}

// Len returns the length of the added objects
func (f *functionContainer) Len() int {
	return f.objects.Len()
}

// Keys returns all the keys in the containers
func (f *functionContainer) Keys() map[string]struct{} {
	keys := make(map[string]struct{}, f.Len())

	for _, key := range f.objects.Keys() {
		stringKey, ok := key.(string)
		if !ok {
			continue
		}

		keys[stringKey] = struct{}{}
	}

	return keys
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *functionContainer) IsInterfaceNil() bool {
	return f == nil
}
