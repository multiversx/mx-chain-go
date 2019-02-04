package interceptor

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// Container is a holder for interceptors organized by type
type Container struct {
	mutex        sync.RWMutex
	interceptors map[string]process.Interceptor
}

// NewContainer will create a new instance of an inteceptor container
func NewContainer() *Container {
	return &Container{
		mutex:        sync.RWMutex{},
		interceptors: make(map[string]process.Interceptor),
	}
}

// Get returns the interceptor stored at a certain key.
//  Returns an error if the element does not exist
func (i *Container) Get(key string) (process.Interceptor, error) {
	i.mutex.RLock()
	interceptor, ok := i.interceptors[key]
	i.mutex.RUnlock()
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}
	return interceptor, nil
}

// Add will add an interceptor at a given key. Returns
//  an error if the element already exists
func (i *Container) Add(key string, interceptor process.Interceptor) error {
	if interceptor == nil {
		return process.ErrNilContainerElement
	}
	i.mutex.Lock()
	defer i.mutex.Unlock()

	_, ok := i.interceptors[key]

	if ok {
		return process.ErrContainerKeyAlreadyExists
	}

	i.interceptors[key] = interceptor
	return nil
}

// Replace will add (or replace if it already exists) an interceptor at a given key
func (i *Container) Replace(key string, interceptor process.Interceptor) error {
	if interceptor == nil {
		return process.ErrNilContainerElement
	}
	i.mutex.Lock()
	i.interceptors[key] = interceptor
	i.mutex.Unlock()
	return nil
}

// Remove will remove an interceptor at a given key
func (i *Container) Remove(key string) {
	i.mutex.Lock()
	delete(i.interceptors, key)
	i.mutex.Unlock()
}

// Len returns the length of the added interceptors
func (i *Container) Len() int {
	i.mutex.RLock()
	l := len(i.interceptors)
	i.mutex.RUnlock()
	return l
}
