package containers

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// InterceptorsContainer is an interceptors holder organized by type
type InterceptorsContainer struct {
	mutex   sync.RWMutex
	objects map[string]process.Interceptor
}

// NewInterceptorsContainer will create a new instance of a container
func NewInterceptorsContainer() *InterceptorsContainer {
	return &InterceptorsContainer{
		mutex:   sync.RWMutex{},
		objects: make(map[string]process.Interceptor),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (ic *InterceptorsContainer) Get(key string) (process.Interceptor, error) {
	ic.mutex.RLock()
	resolver, ok := ic.objects[key]
	ic.mutex.RUnlock()
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}
	return resolver, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (ic *InterceptorsContainer) Add(key string, interceptor process.Interceptor) error {
	if interceptor == nil {
		return process.ErrNilContainerElement
	}
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	_, ok := ic.objects[key]

	if ok {
		return process.ErrContainerKeyAlreadyExists
	}

	ic.objects[key] = interceptor
	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (ic *InterceptorsContainer) Replace(key string, interceptor process.Interceptor) error {
	if interceptor == nil {
		return process.ErrNilContainerElement
	}
	ic.mutex.Lock()
	ic.objects[key] = interceptor
	ic.mutex.Unlock()
	return nil
}

// Remove will remove an object at a given key
func (ic *InterceptorsContainer) Remove(key string) {
	ic.mutex.Lock()
	delete(ic.objects, key)
	ic.mutex.Unlock()
}

// Len returns the length of the added objects
func (ic *InterceptorsContainer) Len() int {
	ic.mutex.RLock()
	l := len(ic.objects)
	ic.mutex.RUnlock()
	return l
}
