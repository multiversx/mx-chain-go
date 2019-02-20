package containers

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// ObjectsContainer is a holder organized by type
type ObjectsContainer struct {
	mutex   sync.RWMutex
	objects map[string]interface{}
}

// NewObjectsContainer will create a new instance of a container
func NewObjectsContainer() *ObjectsContainer {
	return &ObjectsContainer{
		mutex:   sync.RWMutex{},
		objects: make(map[string]interface{}),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (oc *ObjectsContainer) Get(key string) (interface{}, error) {
	oc.mutex.RLock()
	resolver, ok := oc.objects[key]
	oc.mutex.RUnlock()
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}
	return resolver, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (oc *ObjectsContainer) Add(key string, val interface{}) error {
	if val == nil {
		return process.ErrNilContainerElement
	}
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	_, ok := oc.objects[key]

	if ok {
		return process.ErrContainerKeyAlreadyExists
	}

	oc.objects[key] = val
	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (oc *ObjectsContainer) Replace(key string, val interface{}) error {
	if val == nil {
		return process.ErrNilContainerElement
	}
	oc.mutex.Lock()
	oc.objects[key] = val
	oc.mutex.Unlock()
	return nil
}

// Remove will remove an object at a given key
func (oc *ObjectsContainer) Remove(key string) {
	oc.mutex.Lock()
	delete(oc.objects, key)
	oc.mutex.Unlock()
}

// Len returns the length of the added objects
func (oc *ObjectsContainer) Len() int {
	oc.mutex.RLock()
	l := len(oc.objects)
	oc.mutex.RUnlock()
	return l
}
