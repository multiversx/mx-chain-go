package resolver

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// Container is a holder for resolvers organized by type
type Container struct {
	mutex     sync.RWMutex
	resolvers map[string]process.Resolver
}

// NewContainer will create a new instance of a resolver container
func NewContainer() *Container {
	return &Container{
		mutex:     sync.RWMutex{},
		resolvers: make(map[string]process.Resolver),
	}
}

// Get returns the resolver stored at a certain key.
//  Returns an error if the element does not exist
func (i *Container) Get(key string) (process.Resolver, error) {
	i.mutex.RLock()
	resolver, ok := i.resolvers[key]
	i.mutex.RUnlock()
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}
	return resolver, nil
}

// Add will add a resolver at a given key. Returns
//  an error if the element already exists
func (i *Container) Add(key string, resolver process.Resolver) error {
	if resolver == nil {
		return process.ErrNilContainerElement
	}
	i.mutex.Lock()
	defer i.mutex.Unlock()

	_, ok := i.resolvers[key]

	if ok {
		return process.ErrContainerKeyAlreadyExists
	}

	i.resolvers[key] = resolver
	return nil
}

// Replace will add (or replace if it already exists) a resolver at a given key
func (i *Container) Replace(key string, resolver process.Resolver) error {
	if resolver == nil {
		return process.ErrNilContainerElement
	}
	i.mutex.Lock()
	i.resolvers[key] = resolver
	i.mutex.Unlock()
	return nil
}

// Remove will remove a resolver at a given key
func (i *Container) Remove(key string) {
	i.mutex.Lock()
	delete(i.resolvers, key)
	i.mutex.Unlock()
}

// Len returns the length of the added resolvers
func (i *Container) Len() int {
	i.mutex.RLock()
	l := len(i.resolvers)
	i.mutex.RUnlock()
	return l
}
