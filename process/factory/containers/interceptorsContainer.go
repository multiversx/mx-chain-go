package containers

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.InterceptorsContainer = (*interceptorsContainer)(nil)
var log = logger.GetOrCreate("process/factory/containers")

// interceptorsContainer is an interceptors holder organized by type
type interceptorsContainer struct {
	objects *container.MutexMap
}

// NewInterceptorsContainer will create a new instance of a container
func NewInterceptorsContainer() *interceptorsContainer {
	return &interceptorsContainer{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (ic *interceptorsContainer) Get(key string) (process.Interceptor, error) {
	value, ok := ic.objects.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w in interceptors container for key %v", process.ErrInvalidContainerKey, key)
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
	if check.IfNil(interceptor) {
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
		if len(key) == 0 {
			continue
		}

		err := ic.Add(key, interceptors[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (ic *interceptorsContainer) Replace(key string, interceptor process.Interceptor) error {
	if check.IfNil(interceptor) {
		return process.ErrNilContainerElement
	}

	ic.objects.Set(key, interceptor)
	return nil
}

// Remove will remove an object at a given key
func (ic *interceptorsContainer) Remove(key string) {
	ic.objects.Remove(key)
}

// Len returns the length of the added objects
func (ic *interceptorsContainer) Len() int {
	return ic.objects.Len()
}

// Iterate will call the provided handler for each and every key-value pair
func (ic *interceptorsContainer) Iterate(handler func(key string, interceptor process.Interceptor) bool) {
	if handler == nil {
		return
	}

	for _, keyVal := range ic.objects.Keys() {
		key, ok := keyVal.(string)
		if !ok {
			continue
		}

		val, ok := ic.objects.Get(key)
		if !ok {
			continue
		}

		interceptor, ok := val.(process.Interceptor)
		if !ok {
			continue
		}

		shouldContinue := handler(key, interceptor)
		if !shouldContinue {
			return
		}
	}
}

// Close will call the close method on all contained interceptors
func (ic *interceptorsContainer) Close() error {
	var errFound error
	ic.Iterate(func(key string, interceptor process.Interceptor) bool {
		log.Debug("interceptorsContainer closing interceptor", "key", key)
		err := interceptor.Close()
		if err != nil {
			log.Error("error closing interceptor", "key", key, "error", err)
			errFound = err
		}

		return true
	})

	return errFound
}

// IsInterfaceNil returns true if there is no value under the interface
func (ic *interceptorsContainer) IsInterfaceNil() bool {
	return ic == nil
}
