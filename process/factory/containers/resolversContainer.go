package containers

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// ResolversContainer is a resolvers holder organized by type
type ResolversContainer struct {
	mutex   sync.RWMutex
	objects map[string]process.Resolver
}

// NewResolversContainer will create a new instance of a container
func NewResolversContainer() *ResolversContainer {
	return &ResolversContainer{
		mutex:   sync.RWMutex{},
		objects: make(map[string]process.Resolver),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (rc *ResolversContainer) Get(key string) (process.Resolver, error) {
	rc.mutex.RLock()
	resolver, ok := rc.objects[key]
	rc.mutex.RUnlock()
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}
	return resolver, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (rc *ResolversContainer) Add(key string, resolver process.Resolver) error {
	if resolver == nil {
		return process.ErrNilContainerElement
	}
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	_, ok := rc.objects[key]

	if ok {
		return process.ErrContainerKeyAlreadyExists
	}

	rc.objects[key] = resolver
	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (rc *ResolversContainer) Replace(key string, resolver process.Resolver) error {
	if resolver == nil {
		return process.ErrNilContainerElement
	}
	rc.mutex.Lock()
	rc.objects[key] = resolver
	rc.mutex.Unlock()
	return nil
}

// Remove will remove an object at a given key
func (rc *ResolversContainer) Remove(key string) {
	rc.mutex.Lock()
	delete(rc.objects, key)
	rc.mutex.Unlock()
}

// Len returns the length of the added objects
func (rc *ResolversContainer) Len() int {
	rc.mutex.RLock()
	l := len(rc.objects)
	rc.mutex.RUnlock()
	return l
}
